package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

// Endorsement Policy:
// OutOf(2f+1, All orgs...)

type SmartContract struct {
	contractapi.Contract
}

type Event struct {
	ID          string  `json:"id"`
	Lat         float32 `json:"lat"`
	Long        float32 `json:"long"`
	Orientation float32 `json:"orientation"`
	Image       []byte  `json:"image"` // Is this really needed?
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	events := []Event{
		{ID: "asset1", Lat: 36.25, Long: 71.89, Orientation: 78.23, Image: make([]byte, 0)},
	}

	for _, event := range events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(event.ID, eventJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) CreateEvent(ctx contractapi.TransactionContextInterface, id string, lat float32, long float32, orientation float32, image []byte) error {
	exists, err := s.EventExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the event %s already exists", id)
	}

	event := Event{
		ID:          id,
		Lat:         lat,
		Long:        long,
		Orientation: orientation,
		Image:       image,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, eventJSON)
}

func (s *SmartContract) ReadEvent(ctx contractapi.TransactionContextInterface, id string) (*Event, error) {
	eventJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if eventJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var event Event
	err = json.Unmarshal(eventJSON, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *SmartContract) DeleteEvent(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.EventExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) EventExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	eventJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return eventJSON != nil, nil
}

func (s *SmartContract) GetAllEvents(ctx contractapi.TransactionContextInterface) ([]*Event, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var events []*Event
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var event Event
		err = json.Unmarshal(queryResponse.Value, &event)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("", err)
	}
}
