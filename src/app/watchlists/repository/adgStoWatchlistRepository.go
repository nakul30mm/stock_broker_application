package repository

import (
	"context"
	"errors"
	"stock_broker_application/src/models"
	genericModels "stock_broker_application/src/models"
	"watchlists/commons/constants"

	"gorm.io/gorm"
)

type AdgStoWatchlistRepository interface {
	GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*genericModels.User, error)
	DelScripFromWatchlists(ctx context.Context, db *gorm.DB, scripId string, watchlistIds []uint64) ([]uint64, error)
	GetScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string) ([]models.Watchlist, error)
	CheckIfScripExists(ctx context.Context, db *gorm.DB, scripId string) (bool, error)
	GetUsersWatchlists(ctx context.Context, db *gorm.DB, userId uint64, watchlistIds []uint64) ([]uint64, []uint64, error)
	GetWatchlistsWithCapacity(ctx context.Context, db *gorm.DB, belongingWIds []uint64) ([]uint64, []uint64, error)
	GetWatchlistsWithScrip(ctx context.Context, db *gorm.DB, scripId string, addAllowedInWIds []uint64) ([]uint64, []uint64, error)
	AddScripToWatchlists(ctx context.Context, db *gorm.DB, scripId string, AddToWIds []uint64) ([]uint64, error)
	GetWatchlistDetails(ctx context.Context, db *gorm.DB, addedTo []uint64) ([]models.Watchlist, error)
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

// checks if the scripId in the request exists in the scrip_master table
func (repo *adgStoWatchlistsRepository) CheckIfScripExists(ctx context.Context, db *gorm.DB, scripId string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Table("scrip_masters").
		Select("id").
		Where("id = ?", scripId).
		Count(&count).Error

	return count > 0, err
}

// checks if the watchlist ids in the request belongs to the user or not by verifying in watchlists table, returns a list of WIds which belong to the user and a list of WIds which does not
func (repo *adgStoWatchlistsRepository) GetUsersWatchlists(ctx context.Context, db *gorm.DB, userId uint64, watchlistIds []uint64) ([]uint64, []uint64, error) {
	var belongingWIds []uint64

	err := db.WithContext(ctx).
		Table("watchlists").
		Where("user_id = ? AND id IN ?", userId, watchlistIds).
		Pluck("id", &belongingWIds).Error
	if err != nil {
		return nil, nil, err
	}

	belongMap := make(map[uint64]bool)
	for _, b := range belongingWIds {
		belongMap[b] = true
	}

	var notBelongingWIds []uint64
	for _, id := range watchlistIds {
		if !belongMap[id] {
			notBelongingWIds = append(notBelongingWIds, id)
		}
	}
	return belongingWIds, notBelongingWIds, nil
}

// checks if the scripCount < 10 in watchlists table, rertuns a list of Wids with space and a list of WIds with no space
func (repo *adgStoWatchlistsRepository) GetWatchlistsWithCapacity(ctx context.Context, db *gorm.DB, belongingWIds []uint64) ([]uint64, []uint64, error) {
	var haveSpace []uint64
	var fullWIds []uint64

	for _, wId := range belongingWIds {
		var watchlist models.Watchlist

		err := db.WithContext(ctx).
			Table("watchlists").
			Where("id = ? AND scrip_count < ?", wId, 10).
			First(&watchlist).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fullWIds = append(fullWIds, wId)
			} else {
				return nil, nil, err
			}
		} else {
			haveSpace = append(haveSpace, wId)
		}
	}
	return haveSpace, fullWIds, nil
}

// checks if the watchlist already contains the scrip
func (repo *adgStoWatchlistsRepository) GetWatchlistsWithScrip(ctx context.Context, db *gorm.DB, scripId string, haveSpaceWIds []uint64) ([]uint64, []uint64, error) {
	var alreadyExistsIn []uint64
	var addableTo []uint64

	for _, wId := range haveSpaceWIds {
		var watchlistScrip models.WatchlistScrip

		err := db.WithContext(ctx).
			Table("watchlist_scrips").
			Where("watchlist_id = ? AND scrip_id = ?", wId, scripId).
			First(&watchlistScrip).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				addableTo = append(addableTo, wId)
			} else {
				return nil, nil, err
			}
		} else {
			alreadyExistsIn = append(alreadyExistsIn, wId)
		}
	}
	return alreadyExistsIn, addableTo, nil
}

// adds the scripid in the request to the mentioned watchlists,
// returns one []uint64 for list of WIds in which it already exists and another for list of WIds in which it is added
func (repo *adgStoWatchlistsRepository) AddScripToWatchlists(ctx context.Context, db *gorm.DB, scripId string, addableToWIds []uint64) ([]uint64, error) {
	addedTo := []uint64{}

	for _, wId := range addableToWIds {
		insertRecord := models.WatchlistScrip{
			WatchlistId: wId,
			ScripId:     scripId,
		}

		err := db.WithContext(ctx).
			Table("watchlist_scrips").
			Create(&insertRecord).Error

		if err != nil {
			return nil, err
		}

		err1 := db.WithContext(ctx).
			Table("watchlists").
			Where("id = ?", wId).
			Update("scrip_count", gorm.Expr("scrip_count + ?", 1)).Error
		if err1 != nil {
			return nil, err1
		}
		addedTo = append(addedTo, wId)
	}
	return addedTo, nil
}

// returns a list of watchlistWithIds for returning in the service response
// optimised as compared to other helper functions, because here instaed of using a loop, we used IN oprator
func (repo *adgStoWatchlistsRepository) GetWatchlistDetails(ctx context.Context, db *gorm.DB, addedTo []uint64) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist

	err := db.WithContext(ctx).
		Table("watchlists").
		Where("id IN ?", addedTo).
		Find(&watchlists).
		Error

	if err != nil {
		return nil, err
	}

	return watchlists, nil
}

// deleted the scripid in the request from the mentioned watchlists,
// returns a []uint64 for the list of WIds from which a scrip is deleted
func (repo *adgStoWatchlistsRepository) DelScripFromWatchlists(ctx context.Context, db *gorm.DB, scripId string, watchlistIds []uint64) ([]uint64, error) {
	deletedFrom := []uint64{}
	err := db.WithContext(ctx).
		Table("watchlist_scrips").
		Select("watchlist_id").
		Where("scrip_id = ? AND watchlist_id IN ?", scripId, watchlistIds).
		Pluck("watchlist_id", &deletedFrom).Error

	if err != nil {
		return nil, err
	}

	if len(deletedFrom) == 0 {
		return deletedFrom, nil
	}

	err1 := db.WithContext(ctx).
		Table("watchlist_scrips").
		Where("scrip_id = ? AND watchlist_id IN ?", scripId, deletedFrom).
		Delete(&models.WatchlistScrip{}).Error

	if err1 != nil {
		return nil, err1
	}

	err2 := db.WithContext(ctx).
		Table("watchlists").
		Where("id IN ?", deletedFrom).
		Update("scrip_count", gorm.Expr("scrip_count-1")).Error

	if err2 != nil {
		return nil, err
	}

	return deletedFrom, nil
}

// returns the list of watchlists in which the scrip is present,
// returns a []uint64 for the list of WIds in which the scrip is present
func (repo *adgStoWatchlistsRepository) GetScripFromWatchlists(ctx context.Context, db *gorm.DB, userId uint64, scripId string) ([]models.Watchlist, error) {
	var watchlists []models.Watchlist

	err := db.WithContext(ctx).
		Table("watchlists AS w").
		Select("w.id, w.watchlist_name").
		Joins("JOIN watchlist_scrips ws ON w.id = ws.watchlist_id").
		Where("w.user_id = ? AND ws.scrip_id = ?", userId, scripId).
		Find(&watchlists).Error

	if err != nil {
		return nil, err
	}

	return watchlists, nil
}
