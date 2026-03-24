package business

import (
	"context"
	"errors"
	"fmt"
	"stock_broker_application/src/utils"
	"strings"
	"watchlists/commons/constants"
	"watchlists/models"
	"watchlists/repository"

	"gorm.io/gorm"
)

type AdgStoWatchlistService struct {
	adgStoWatchlistRepository repository.AdgStoWatchlistRepository
}

func NewadgStoWatchlistService(adgStoWatchlistRepository repository.AdgStoWatchlistRepository) *AdgStoWatchlistService {
	return &AdgStoWatchlistService{
		adgStoWatchlistRepository: adgStoWatchlistRepository,
	}
}

func (service *AdgStoWatchlistService) AdgStoWatchlist(ctx context.Context, username string, request models.BffAdgStoWatchlistRequest) ([]string, []models.WatchlistWithId, error) {
	postgresClient := utils.GetPostgresClient().GormDB

	ReqAction := models.Actiontype(strings.ToUpper(string(request.Action)))

	user, err := service.adgStoWatchlistRepository.GetUserByUsername(ctx, postgresClient, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, constants.UserNotFoundError
		}
		return nil, nil, err
	}

	switch ReqAction {
	case models.AddAction:
		// warnings := []string{}
		// respWatchlistsWithIds := []models.WatchlistWithId{}

		// watchlists, finalIds, full, existing, err := service.adgStoWatchlistRepository.AddScripToWatchlists(ctx, postgresClient, user.ID, request.ScripId, request.WatchlistIds)
		// if err != nil {
		// 	return warnings, respWatchlistsWithIds, err
		// }

		// if len(watchlists) == 0 {
		// 	return warnings, respWatchlistsWithIds, constants.InvalidWatchlistsError
		// }

		// if len(finalIds) == 0 {
		// 	return warnings, respWatchlistsWithIds, constants.AllWatchlistsFullError
		// }

		// if len(full) > 0 {
		// 	warnings = append(warnings, fmt.Sprintf("watchlistIds %v already have 10 scrips", full))
		// }

		// if len(existing) > 0 {
		// 	warnings = append(warnings, fmt.Sprintf("watchlistIds %v already have scripId %s", existing, request.ScripId))
		// }

		// validMap := make(map[uint64]bool)
		// for _, w := range watchlists {
		// 	validMap[w.Id] = true
		// }

		// invalid := []uint64{}
		// for _, id := range request.WatchlistIds {
		// 	if !validMap[id] {
		// 		invalid = append(invalid, id)
		// 	}
		// }

		// if len(invalid) > 0 {
		// 	warnings = append(warnings, fmt.Sprintf("watchlistIds %v are invalid", invalid))
		// }

		// resp, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, finalIds)
		// if err != nil {
		// 	return nil, nil, err
		// }
		// for _, r := range resp {
		// 	respWatchlistsWithIds = append(respWatchlistsWithIds, models.WatchlistWithId{
		// 		Id:   r.Id,
		// 		Name: r.WatchlistName,
		// 	})
		// }
		// return warnings, respWatchlistsWithIds, nil

		warnings := []string{}
		respWatchlistWithIds := []models.WatchlistWithId{}

		addedIds, err := service.adgStoWatchlistRepository.AddScripToWatchlistss(ctx, postgresClient, user.ID, request.ScripId, request.WatchlistIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		if len(addedIds) == 0 { //dont append to warning, return error.
			// warnings = append(warnings, "scrip was not added to any watchlists - maybe full, already contain the scrip or invalid")
			return warnings, respWatchlistWithIds, constants.ScripNotAddedToAnyWatchlistsError
		}
		if len(addedIds) < len(request.WatchlistIds) {
			warnings = append(warnings, "scrip was not added to some watchlists")
		}

		resp, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, addedIds)
		if err != nil {
			return warnings, respWatchlistWithIds, constants.DatabaseQueryError
		}
		for _, r := range resp {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   r.Id,
				Name: r.WatchlistName,
			})
		}

		return warnings, respWatchlistWithIds, nil

	case models.DelAction:
		warnings := []string{}
		respWatchlistWithIds := []models.WatchlistWithId{}

		userWatchlists, err := service.adgStoWatchlistRepository.DelScripFromWatchlists(ctx, postgresClient, user.ID, request.ScripId, request.WatchlistIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if len(userWatchlists) == 0 {
			return warnings, respWatchlistWithIds, constants.InvalidWatchlistsError
		}

		//check for invalid ids for warnings
		usersIdsMap := make(map[uint64]bool)
		inValidIds := []uint64{}

		for _, id := range userWatchlists {
			usersIdsMap[id] = true
		}

		for _, id := range request.WatchlistIds {
			if !usersIdsMap[id] {
				inValidIds = append(inValidIds, id)
			}
		}

		if len(inValidIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds %v do not belong to the user", inValidIds))
		}

		response, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, userWatchlists)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		for _, w := range response {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		return warnings, respWatchlistWithIds, nil

	case models.GetAction:
		warnings := []string{}
		respWatchlistWithId := []models.WatchlistWithId{}

		watchlists, err := service.adgStoWatchlistRepository.GetScripFromWatchlists(ctx, postgresClient, user.ID, request.ScripId)
		if err != nil {
			return warnings, respWatchlistWithId, err
		}
		if len(watchlists) == 0 {
			return warnings, respWatchlistWithId, constants.ScripNotInWatchlistsError
		}

		for _, w := range watchlists {
			respWatchlistWithId = append(respWatchlistWithId, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		return warnings, respWatchlistWithId, nil

	default:
		return nil, nil, constants.InvalidActionTypeError
	}
}
