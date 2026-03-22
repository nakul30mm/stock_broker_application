package repository

import (
	"context"
	"errors"
	"stock_broker_application/src/models"
	genericModels "stock_broker_application/src/models"
	"watchlists/commons/constants"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdgStoWatchlistRepository interface {
	GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*genericModels.User, error)
	GetWatchlistDetails(ctx context.Context, db *gorm.DB, addedTo []uint64) ([]models.Watchlist, error)
	DelScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, []uint64, error)
	GetScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string) ([]models.Watchlist, error)
	AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]models.Watchlist, []uint64, []uint64, []uint64, error)
}

type adgStoWatchlistsRepository struct{}

func NewadgStoWatchlistsRepoitory() *adgStoWatchlistsRepository {
	return &adgStoWatchlistsRepository{}
}

// checks if the user exists in the given table or not
func (repo *adgStoWatchlistsRepository) GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*genericModels.User, error) {
	var user genericModels.User

	result := db.WithContext(ctx).
		Table(constants.UsersTableName).
		Where(constants.Username, username).
		First(&user) //change table name
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// returns a list of watchlistWithIds for returning in the service response
// optimised as compared to other helper functions, because here instaed of using a loop, we used IN oprator
func (repo *adgStoWatchlistsRepository) GetWatchlistDetails(ctx context.Context, db *gorm.DB, addedTo []uint64) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist

	err := db.WithContext(ctx).
		Table("watchlists").
		Where("id IN ?", addedTo).
		Find(&watchlists).Error

	if err != nil {
		return nil, constants.DatabaseQueryError
	}

	return watchlists, nil
}

// returns the list of watchlists in which the scrip is present,
// returns a []uint64 for the list of WIds in which the scrip is present
func (repo *adgStoWatchlistsRepository) GetScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist

	err := db.WithContext(ctx).
		Table(constants.WatchlistsTableName+" AS w").
		// Select("w.id, w.watchlist_name").
		Select("w."+constants.FieldId+", w."+constants.FieldWatchlistName).
		// Joins("JOIN watchlist_scrips ws ON w.id = ws.watchlist_id").
		Joins("JOIN "+constants.WatchlistScripsTableName+" AS ws ON w."+constants.FieldId+" = ws."+constants.FieldWatchlistId).
		// Where("w.user_id = ? AND ws.scrip_id = ?", userId, scripId).
		Where("w."+constants.FieldUserId+" = ? AND ws."+constants.FieldScripId+" = ?", userId, scripId).
		Find(&watchlists).Error
	if err != nil {
		return nil, constants.DatabaseQueryError
	}

	return watchlists, nil
}

// deletes the scripid in the request from the mentioned watchlists,
// returns a []uint64 for the list of WIds from which a scrip is deleted
func (repo *adgStoWatchlistsRepository) DelScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, []uint64, error) {
	// 	if len(watchlistIds) == 0 {
	// 		return []uint64{}, nil
	// 	}
	userWatchlists := []uint64{}
	err := db.WithContext(ctx).
		Table(constants.WatchlistsTableName).
		Where(constants.FieldUserId+" = ? AND "+constants.FieldId+" IN ?", userId, watchlistIds).
		Pluck(constants.FieldId, &userWatchlists).Error
	if err != nil {
		return nil, nil, constants.DatabaseQueryError
	}
	if len(userWatchlists) == 0 {
		return userWatchlists, []uint64{}, nil
	}

	deletedRecords := []models.WatchlistScrip{}
	err = db.WithContext(ctx).
		Table(constants.WatchlistScripsTableName).
		Clauses(clause.Returning{}).
		Where(constants.FieldScripId+" = ? AND "+constants.FieldWatchlistId+" IN ?", scripId, userWatchlists).
		Delete(&deletedRecords).Error
	if err != nil {
		return nil, nil, constants.DatabaseQueryError
	}

	deletedFrom := []uint64{}
	for _, rec := range deletedRecords {
		deletedFrom = append(deletedFrom, rec.WatchlistId)
	}
	if len(deletedFrom) == 0 {
		return userWatchlists, deletedFrom, nil
	}

	err = db.WithContext(ctx).
		Table(constants.WatchlistsTableName).
		Where(constants.FieldId+" IN ?", deletedFrom).
		Updates(map[string]interface{}{
			constants.FieldScripCount:    gorm.Expr("GREATEST(scrip_count - 1, 0)"),
			constants.FieldLastUpdatedAt: gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, nil, constants.DatabaseQueryError
	}

	return userWatchlists, deletedFrom, nil
}

func (repo *adgStoWatchlistsRepository) AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]models.Watchlist, []uint64, []uint64, []uint64, error) {
	//chck for existwnce of scrip
	var count int64
	err := db.WithContext(ctx).
		Table(constants.ScripMastersTableName).
		Where(constants.FieldId+" = ?", scripId).
		Count(&count).Error
	if err != nil {
		return nil, nil, nil, nil, constants.DatabaseQueryError
	}
	if count == 0 {
		return nil, nil, nil, nil, constants.ScripDoesNotExistError
	}

	//fetching user's watchlists
	var watchlists []models.Watchlist
	err = db.WithContext(ctx).
		Table(constants.WatchlistsTableName).
		Select(constants.FieldId, constants.FieldWatchlistName, constants.FieldScripCount).
		Where(constants.FieldUserId+" = ? AND "+constants.FieldId+" IN ?", userId, watchlistIds).
		Find(&watchlists).Error
	if err != nil {
		return nil, nil, nil, nil, constants.DatabaseQueryError
	}

	if len(watchlists) == 0 {
		return watchlists, nil, nil, nil, nil
	}

	//checking capacity
	var elligible []uint64
	var full []uint64

	for _, w := range watchlists {
		if w.ScripCount >= 10 {
			full = append(full, w.Id)
		} else {
			elligible = append(elligible, w.Id)
		}
	}

	//checking duplicates
	var existingIds []uint64

	err = db.WithContext(ctx).
		Table(constants.WatchlistScripsTableName).
		Where(constants.FieldWatchlistId+" IN ? AND "+constants.FieldScripId+" = ?", elligible, scripId).
		Pluck(constants.FieldWatchlistId, &existingIds).Error
	if err != nil {
		return nil, nil, nil, nil, constants.DatabaseQueryError
	}

	existsMap := make(map[uint64]bool)
	for _, id := range existingIds {
		existsMap[id] = true
	}

	//inserting list
	var finalIds []uint64

	for _, id := range elligible {
		if !existsMap[id] {
			finalIds = append(finalIds, id)
		}
	}

	//insertion
	var inserts []models.WatchlistScrip
	for _, id := range finalIds {
		inserts = append(inserts, genericModels.WatchlistScrip{
			WatchlistId: id,
			ScripId:     scripId,
		})
	}
	if len(inserts) > 0 {
		err := db.WithContext(ctx).
			Table(constants.WatchlistScripsTableName).
			Create(&inserts).Error
		if err != nil {
			return nil, nil, nil, nil, constants.DatabaseQueryError
		}

		//updating the count and timestamp
		err = db.WithContext(ctx).
			Table(constants.WatchlistsTableName).
			Where(constants.FieldId+" IN ?", finalIds).
			Updates(map[string]interface{}{
				constants.FieldScripCount:    gorm.Expr("scrip_count + 1"),
				constants.FieldLastUpdatedAt: gorm.Expr("NOW()"),
			}).Error

		if err != nil {
			return nil, nil, nil, nil, constants.DatabaseQueryError
		}

	}

	return watchlists, finalIds, full, existingIds, nil
}
