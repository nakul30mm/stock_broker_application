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
		//check if scripid exists, if not give err and return
		exists, err := service.adgStoWatchlistRepository.CheckIfScripExists(ctx, postgresClient, request.ScripId)
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			return nil, nil, errors.New(constants.ErrScripDoesnotExist)
		}

		var warnings []string

		//check if all the mentioned watchlists belongs to the user, if not continue to check next watchlist
		belongingWIds, notBelongingWIds, err := service.adgStoWatchlistRepository.GetUsersWatchlists(ctx, postgresClient, user.ID, request.WatchlistIds)
		if err != nil {
			return nil, nil, err
		}
		if len(notBelongingWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds: %v, do not belong to the user", notBelongingWIds))
		}

		//check scripCount in those watchlists, if >= 10, give warning and continue for adding to next watchlist
		haveSpaceWIds, fullWIds, err := service.adgStoWatchlistRepository.GetWatchlistsWithCapacity(ctx, postgresClient, belongingWIds)
		if err != nil {
			return nil, nil, err
		}
		if len(fullWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds: %v, already have 10 scrips", fullWIds))
		}

		//check if that scrip already exists in the watchlist, if not add, else add to warning
		alreadyIn, addableTo, err := service.adgStoWatchlistRepository.GetWatchlistsWithScrip(ctx, postgresClient, request.ScripId, haveSpaceWIds)
		if err != nil {
			return nil, nil, err
		}
		if len(alreadyIn) > 0 {
			warnings = append(warnings, fmt.Sprintf("scripId %s already exists in WatchlistIds: %v", request.ScripId, alreadyIn))
		}
		if len(addableTo) == 0 {
			return warnings, nil, nil
		}

		//add the scrip in the request to the watchlists ini the request
		addedTo, err := service.adgStoWatchlistRepository.AddScripToWatchlists(ctx, postgresClient, request.ScripId, addableTo)
		if err != nil {
			return nil, nil, err
		}

		watchlists, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, addedTo)
		if err != nil {
			return nil, nil, err
		}

		respWatchlistWithIds := []models.WatchlistWithId{} //if did using var, and if it remains empty, json returns null, but if we initialize it, json returns []

		for _, wtchlst := range watchlists {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   wtchlst.Id,
				Name: wtchlst.WatchlistName,
			})
		}

		return warnings, respWatchlistWithIds, nil

	case "del":
		//checking if the watchlists in the request belongs to the user
		var warnings []string

		belongingWIds, notBelongingWIds, err := service.adgStoWatchlistRepository.GetUsersWatchlists(ctx, postgresClient, user.ID, request.WatchlistIds)
		if err != nil {
			return nil, nil, err
		}
		if len(notBelongingWIds) > 0 {
			warnings = append(warnings, fmt.Sprintf("watchlistIds %v does not belong to the user", notBelongingWIds))
		}

		if len(belongingWIds) == 0 {
			return nil, []models.WatchlistWithId{}, nil
		}

		deletedFrom, err := service.adgStoWatchlistRepository.DelScripFromWatchlists(ctx, postgresClient, request.ScripId, belongingWIds)
		if err != nil {
			return nil, nil, err
		}

		if len(deletedFrom) == 0 {
			warnings = append(warnings, fmt.Sprintf("scripId %s deleted from watchlistIds %v", request.ScripId, request.WatchlistIds))
		}

		watchlists, err := service.adgStoWatchlistRepository.GetWatchlistDetails(ctx, postgresClient, deletedFrom)
		if err != nil {
			return nil, nil, err
		}

		respWatchlistWithIds := []models.WatchlistWithId{}
		for _, w := range watchlists {
			respWatchlistWithIds = append(respWatchlistWithIds, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		return warnings, respWatchlistWithIds, nil

	case "get":
		watchlists, err := service.adgStoWatchlistRepository.GetScripFromWatchlists(ctx, postgresClient, user.ID, request.ScripId)
		if err != nil {
			return nil, nil, err
		}

		respWatchlistWithId := []models.WatchlistWithId{}
		for _, w := range watchlists {
			respWatchlistWithId = append(respWatchlistWithId, models.WatchlistWithId{
				Id:   w.Id,
				Name: w.WatchlistName,
			})
		}
		if len(respWatchlistWithId) == 0 {
			return []string{"scrip not found in any watchlist"}, respWatchlistWithId, nil
		}
		return nil, respWatchlistWithId, nil

	default:
		return nil, nil, errors.New(constants.ErrInvalidActiontype)
	}
}

//either return on the basis of adg functions or return a struct witn necessary fields
