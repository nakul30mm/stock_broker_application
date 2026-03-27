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
	AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, []uint64, error)
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

func (repo *adgStoWatchlistsRepository) AddScripToWatchlistss(ctx context.Context, db *gorm.DB, userId uint64, scripId string, watchlistIds []uint64) ([]uint64, []uint64, error) {

	type Result struct {
		ValidWatchlistIds    uint64  `gorm:"column:valid_watchlist_ids"`
		InsertedWatchlistIds *uint64 `gorm:"column:inserted_watchlist_ids"` //ptr because all vw.id will appear, but ir.watchlist_id maybe null, go has no null for uint64,
		// so when took ptr, corresponding value for null becomes nil in go
	}

	results := []Result{}
	valid := []uint64{}
	inserted := []uint64{}

	query := `
	WITH valid_watchlists AS (
			SELECT id, scrip_count
			FROM watchlists w
			WHERE w.id IN (?)
			AND w.user_id = ?
		),
		insert_records AS (
			INSERT INTO watchlist_scrips (watchlist_id, scrip_id)
			SELECT vw.id, ?
			FROM valid_watchlists vw
			WHERE vw.scrip_count < 10
			ON CONFLICT (watchlist_id, scrip_id) DO NOTHING
			RETURNING watchlist_id  
		),
		update_count AS (
			UPDATE watchlists
			SET scrip_count = scrip_count + 1,
				last_updated_at = NOW()
			WHERE id IN (SELECT watchlist_id FROM insert_records)
		)
		SELECT vw.id AS valid_watchlist_ids,
			ir.watchlist_id AS inserted_watchlist_ids
		FROM valid_watchlists AS vw
		LEFT JOIN insert_records AS ir
			ON vw.id = ir.watchlist_id;
	`
	err := db.WithContext(ctx).Raw(query, watchlistIds, userId, scripId).Scan(&results).Error
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
			return nil, nil, constants.ScripDoesNotExistError
		}
		return nil, nil, err
	}

	for _, res := range results {
		valid = append(valid, res.ValidWatchlistIds)
		if res.InsertedWatchlistIds != nil {
			inserted = append(inserted, *res.InsertedWatchlistIds)
		}
	}

	return valid, inserted, nil
}
