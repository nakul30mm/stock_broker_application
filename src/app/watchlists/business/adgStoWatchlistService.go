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

	ReqAction := strings.ToLower(string(request.Action))

	user, err := service.adgStoWatchlistRepository.GetUserByUsername(ctx, postgresClient, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errors.New(constants.ErrUserNotFound)
		}
		return nil, nil, err
	}

	switch ReqAction {
	case "add":
		var warnings []string                              //for returning warnings
		respWatchlistWithIds := []models.WatchlistWithId{} //if did using var, and if it remains empty, json returns null, but if we initialize it, json returns []

		//check if scripid exists, if not give err and return
		exists, err := service.adgStoWatchlistRepository.CheckIfScripExists(ctx, postgresClient, request.ScripId)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if !exists {
			return warnings, respWatchlistWithIds, errors.New(constants.ErrScripDoesnotExist)
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

	case "del":
		//checking if the watchlists in the request belongs to the user
		var warnings []string
		respWatchlistWithIds := []models.WatchlistWithId{}

		belongingWIds, notBelongingWIds, err := service.adgStoWatchlistRepository.GetUsersWatchlists(ctx, postgresClient, user.ID, request.WatchlistIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}
		if len(notBelongingWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds %v does not belong to the user", notBelongingWIds))
		}

		if len(belongingWIds) == 0 {
			return warnings, respWatchlistWithIds, nil
		}

		deletedFrom, err := service.adgStoWatchlistRepository.DelScripFromWatchlists(ctx, postgresClient, request.ScripId, belongingWIds)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		if len(deletedFrom) == 0 {
			warnings = append(warnings, fmt.Sprintf("scripId %s deleted from watchlistIds %v", request.ScripId, request.WatchlistIds))
		}

		watchlists, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, deletedFrom)
		if err != nil {
			return warnings, respWatchlistWithIds, err
		}

		for _, w := range watchlists {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		return warnings, respWatchlistWithIds, nil

	case "get":
		warnings := []string{}
		respWatchlistWithId := []models.WatchlistWithId{}

		watchlists, err := service.adgStoWatchlistRepository.GetScripFromWatchlists(ctx, postgresClient, user.ID, request.ScripId)
		if err != nil {
			return warnings, respWatchlistWithId, err
		}

		for _, w := range watchlists {
			respWatchlistWithId = append(respWatchlistWithId, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		if len(respWatchlistWithId) == 0 {
			return []string{"scrip not found in any watchlist"}, respWatchlistWithId, nil
		}
		return warnings, respWatchlistWithId, nil

	default:
		return nil, nil, errors.New(constants.ErrInvalidActiontype)
	}
}

//either return on the basis of adg functions or return a struct witn necessary fields
