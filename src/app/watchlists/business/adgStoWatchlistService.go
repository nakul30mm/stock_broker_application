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
		warnings := []string{}                             //for returning warnings
		respWatchlistWithIds := []models.WatchlistWithId{} //if did using var, and if it remains empty, json returns null, but if we initialize it, json returns []

		//check if scripid exists, if not give err and return
		exists, err := service.adgStoWatchlistRepository.CheckIfScripExists(ctx, postgresClient, request.ScripId)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if !exists {
			return warnings, respWatchlistWithIds, constants.ScripDoesNotExistError
		}

		//check if all the mentioned watchlists belongs to the user, if not continue to check next watchlist
		belongingWIds, notBelongingWIds, err := service.adgStoWatchlistRepository.GetUsersWatchlists(ctx, postgresClient, user.ID, request.WatchlistIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if len(notBelongingWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds: %v, do not belong to the user", notBelongingWIds))
		}

		//check scripCount in those watchlists, if >= 10, give warning and continue for adding to next watchlist
		haveSpaceWIds, fullWIds, err := service.adgStoWatchlistRepository.GetWatchlistsWithCapacity(ctx, postgresClient, belongingWIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if len(fullWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds: %v, already have 10 scrips", fullWIds))
		}

		alreadyIn, addableTo, err := service.adgStoWatchlistRepository.GetWatchlistsWithScrip(ctx, postgresClient, request.ScripId, haveSpaceWIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if len(alreadyIn) > 0 {
			warnings = append(warnings, fmt.Sprintf("scripId %s already exists in WatchlistIds: %v", request.ScripId, alreadyIn))
		}
		if len(addableTo) == 0 {
			return warnings, respWatchlistWithIds, nil
		}

		addedTo, err := service.adgStoWatchlistRepository.AddScripToWatchlists(ctx, postgresClient, request.ScripId, addableTo)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		watchlistDetails, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, addedTo)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		for _, w := range watchlistDetails {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}

		return warnings, respWatchlistWithIds, nil

	case models.DelAction:
		warnings := []string{}
		respWatchlistWithIds := []models.WatchlistWithId{}

		deletedFrom, err := service.adgStoWatchlistRepository.DelScripFromWatchlists(ctx, postgresClient, user.ID, request.ScripId, request.WatchlistIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		if len(deletedFrom) == 0 {
			return warnings, respWatchlistWithIds, constants.ScripNotInWatchlistsError
		}

		if len(deletedFrom) != len(request.WatchlistIds) {
			warnings = append(warnings, "some watchlists were invalid or did not contain the scrip")
		}

		response, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, deletedFrom)
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
