package repository

import (
	"context"
	"errors"
	"stock_broker_application/src/models"
	genericModels "stock_broker_application/src/models"
	"strings"
	"watchlists/commons/constants"

	"gorm.io/gorm"
)

type AdgStoWatchlistRepository interface {
	GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*genericModels.User, error)
	GetWatchlistDetails(ctx context.Context, db *gorm.DB, addedTo []uint64) ([]models.Watchlist, error)
	DelScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, error)
	GetScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string) ([]models.Watchlist, error)
	// AddScripToWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]models.Watchlist, []uint64, []uint64, []uint64, error)
	AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, error)
}

type adgStoWatchlistsRepository struct{}

func NewadgStoWatchlistsRepository() *adgStoWatchlistsRepository {
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
// returns a []uint64 for the list of WIds which belong to the user.
func (repo *adgStoWatchlistsRepository) DelScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, error) {
	if len(watchlistIds) == 0 {
		return []uint64{}, nil
	}

	// validWatchlists := []uint64{}
	// err := db.WithContext(ctx).
	// 	Table(constants.WatchlistsTableName).
	// 	Where(constants.FieldUserId+" = ? AND "+constants.FieldId+" IN ?", userId, watchlistIds).
	// 	Pluck(constants.FieldId, &validWatchlists).Error
	// if err != nil {
	// 	return nil, constants.DatabaseQueryError
	// }
	// if len(validWatchlists) == 0 {
	// 	return validWatchlists, nil
	// }

	// deletedRecords := []models.WatchlistScrip{}
	// err = db.WithContext(ctx).
	// 	Table(constants.WatchlistScripsTableName).
	// 	Clauses(clause.Returning{}).
	// 	Where(constants.FieldScripId+" = ? AND "+constants.FieldWatchlistId+" IN ?", scripId, validWatchlists).
	// 	Delete(&deletedRecords).Error
	// if err != nil {
	// 	return nil, constants.DatabaseQueryError
	// }

	// deletedFrom := []uint64{}
	// for _, rec := range deletedRecords {
	// 	deletedFrom = append(deletedFrom, rec.WatchlistId)
	// }

	// if len(deletedFrom) > 0 {
	// 	err = db.WithContext(ctx).
	// 		Table(constants.WatchlistsTableName).
	// 		Where(constants.FieldId+" IN ?", deletedFrom).
	// 		Updates(map[string]interface{}{
	// 			constants.FieldScripCount:    gorm.Expr("GREATEST(scrip_count - 1, 0)"),
	// 			constants.FieldLastUpdatedAt: gorm.Expr("NOW()"),
	// 		}).Error
	// 	if err != nil {
	// 		return nil, constants.DatabaseQueryError
	// 	}
	// }
	query := `
	WITH valid_watchlists AS(
		SELECT id 
		FROM watchlists 
		WHERE user_id = ? 
		AND id IN (?)
	),
	deleted_from AS (
		DELETE FROM watchlist_scrips
		WHERE scrip_id = ?
		AND watchlist_id IN (SELECT id FROM valid_watchlists)
		RETURNING watchlist_id
	),
	update_count AS (
		UPDATE watchlists
		SET scrip_count = GREATEST(scrip_count - 1, 0),
			last_updated_at = NOW()
		WHERE id IN (SELECT watchlist_id FROM deleted_from)
	)
	SELECT id FROM valid_watchlists
	`
	var validWatchlists []uint64
	err := db.WithContext(ctx).Raw(query, userId, watchlistIds, scripId).Scan(&validWatchlists).Error
	if err != nil {
		return nil, constants.DatabaseQueryError
	}

	return validWatchlists, nil
}

// func (repo *adgStoWatchlistsRepository) AddScripToWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]models.Watchlist, []uint64, []uint64, []uint64, error) {
// 	//chck for existwnce of scrip
// 	var count int64
// 	err := db.WithContext(ctx).
// 		Table(constants.ScripMastersTableName).
// 		Where(constants.FieldId+" = ?", scripId).
// 		Count(&count).Error
// 	if err != nil {
// 		return nil, nil, nil, nil, constants.DatabaseQueryError
// 	}
// 	if count == 0 {
// 		return nil, nil, nil, nil, constants.ScripDoesNotExistError
// 	}

// 	//fetching user's watchlists
// 	var watchlists []models.Watchlist
// 	err = db.WithContext(ctx).
// 		Table(constants.WatchlistsTableName).
// 		Select(constants.FieldId, constants.FieldWatchlistName, constants.FieldScripCount).
// 		Where(constants.FieldUserId+" = ? AND "+constants.FieldId+" IN ?", userId, watchlistIds).
// 		Find(&watchlists).Error
// 	if err != nil {
// 		return nil, nil, nil, nil, constants.DatabaseQueryError
// 	}

// 	if len(watchlists) == 0 {
// 		return watchlists, nil, nil, nil, nil
// 	}

// 	//checking capacity
// 	var elligible []uint64
// 	var full []uint64

// 	for _, w := range watchlists {
// 		if w.ScripCount >= 10 {
// 			full = append(full, w.Id)
// 		} else {
// 			elligible = append(elligible, w.Id)
// 		}
// 	}

// 	//checking duplicates
// 	var existingIds []uint64

// 	err = db.WithContext(ctx).
// 		Table(constants.WatchlistScripsTableName).
// 		Where(constants.FieldWatchlistId+" IN ? AND "+constants.FieldScripId+" = ?", elligible, scripId).
// 		Pluck(constants.FieldWatchlistId, &existingIds).Error
// 	if err != nil {
// 		return nil, nil, nil, nil, constants.DatabaseQueryError
// 	}

// 	existsMap := make(map[uint64]bool)
// 	for _, id := range existingIds {
// 		existsMap[id] = true
// 	}

// 	//inserting list
// 	var finalIds []uint64

// 	for _, id := range elligible {
// 		if !existsMap[id] {
// 			finalIds = append(finalIds, id)
// 		}
// 	}

// 	//insertion
// 	var inserts []models.WatchlistScrip
// 	for _, id := range finalIds {
// 		inserts = append(inserts, genericModels.WatchlistScrip{
// 			WatchlistId: id,
// 			ScripId:     scripId,
// 		})
// 	}
// 	if len(inserts) > 0 {
// 		err := db.WithContext(ctx).
// 			Table(constants.WatchlistScripsTableName).
// 			Create(&inserts).Error
// 		if err != nil {
// 			return nil, nil, nil, nil, constants.DatabaseQueryError
// 		}

// 		//updating the count and timestamp
// 		err = db.WithContext(ctx).
// 			Table(constants.WatchlistsTableName).
// 			Where(constants.FieldId+" IN ?", finalIds).
// 			Updates(map[string]interface{}{
// 				constants.FieldScripCount:    gorm.Expr("scrip_count + 1"),
// 				constants.FieldLastUpdatedAt: gorm.Expr("NOW()"),
// 			}).Error

// 		if err != nil {
// 			return nil, nil, nil, nil, constants.DatabaseQueryError
// 		}
// 	}

// return watchlists, finalIds, full, existingIds, nil
// }

func (repo *adgStoWatchlistsRepository) AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, error) {
	var addedIds []uint64
	err := db.WithContext(ctx).Raw(`
		INSERT INTO watchlist_scrips (watchlist_id, scrip_id) 
		SELECT w.id, ? 
		FROM watchlists w 
		WHERE w.user_id = ? 
		AND w.id IN ? 
		AND w.scrip_count < ? 
		ON CONFLICT (watchlist_id, scrip_id) DO NOTHING 
		RETURNING watchlist_id`,
		scripId, userId, watchlistIds, 10).
		Scan(&addedIds).Error
	if err != nil {
		// if errors.Is(err, gorm.ErrForeignKeyViolated) { //or if did not work- strings.contains(err, "foreign key")
		// 	return addedIds, constants.ScripDoesNotExistError
		// }
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
			return addedIds, constants.ScripDoesNotExistError
		}
	}

	if len(addedIds) > 0 {
		err := db.WithContext(ctx).Table(constants.WatchlistsTableName).Where(constants.FieldId+" IN ?", addedIds).Updates(map[string]interface{}{
			constants.FieldScripCount:    gorm.Expr(constants.FieldScripCount + " + 1"),
			constants.FieldLastUpdatedAt: gorm.Expr("NOW()"),
		}).Error
		if err != nil {
			return addedIds, constants.DatabaseQueryError
		}
	}
	return addedIds, nil
}
