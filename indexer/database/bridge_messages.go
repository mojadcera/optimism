package database

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"

	"github.com/google/uuid"
)

/**
 * Types
 */

type BridgeMessage struct {
	MessageHash common.Hash `gorm:"primaryKey;serializer:json"`
	Nonce       U256

	SentMessageEventGUID    uuid.UUID
	RelayedMessageEventGUID *uuid.UUID

	Tx       Transaction `gorm:"embedded"`
	GasLimit U256
}

type L1BridgeMessage struct {
	BridgeMessage         `gorm:"embedded"`
	TransactionSourceHash common.Hash `gorm:"serializer:json"`
}

type L2BridgeMessage struct {
	BridgeMessage             `gorm:"embedded"`
	TransactionWithdrawalHash common.Hash `gorm:"serializer:json"`
}

type BridgeMessagesView interface {
	L1BridgeMessage(common.Hash) (*L1BridgeMessage, error)
	L1BridgeMessageWithFilter(BridgeMessage) (*L1BridgeMessage, error)

	L2BridgeMessage(common.Hash) (*L2BridgeMessage, error)
	L2BridgeMessageWithFilter(BridgeMessage) (*L2BridgeMessage, error)
}

type BridgeMessagesDB interface {
	BridgeMessagesView

	StoreL1BridgeMessages([]L1BridgeMessage) error
	MarkRelayedL1BridgeMessage(common.Hash, uuid.UUID) error

	StoreL2BridgeMessages([]L2BridgeMessage) error
	MarkRelayedL2BridgeMessage(common.Hash, uuid.UUID) error
}

/**
 * Implementation
 */

type bridgeMessagesDB struct {
	gorm *gorm.DB
}

func newBridgeMessagesDB(db *gorm.DB) BridgeMessagesDB {
	return &bridgeMessagesDB{gorm: db}
}

/**
 * Arbitrary Messages Sent from L1
 */

func (db bridgeMessagesDB) StoreL1BridgeMessages(messages []L1BridgeMessage) error {
	result := db.gorm.Create(&messages)
	return result.Error
}

func (db bridgeMessagesDB) L1BridgeMessage(msgHash common.Hash) (*L1BridgeMessage, error) {
	return db.L1BridgeMessageWithFilter(BridgeMessage{MessageHash: msgHash})
}

func (db bridgeMessagesDB) L1BridgeMessageWithFilter(filter BridgeMessage) (*L1BridgeMessage, error) {
	var sentMessage L1BridgeMessage
	result := db.gorm.Where(&filter).Take(&sentMessage)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &sentMessage, nil
}

func (db bridgeMessagesDB) MarkRelayedL1BridgeMessage(messageHash common.Hash, relayEvent uuid.UUID) error {
	message, err := db.L1BridgeMessage(messageHash)
	if err != nil {
		return err
	} else if message == nil {
		return fmt.Errorf("L1BridgeMessage with message hash %s not found", messageHash)
	}

	message.RelayedMessageEventGUID = &relayEvent
	result := db.gorm.Save(message)
	return result.Error
}

/**
 * Arbitrary Messages Sent from L2
 */

func (db bridgeMessagesDB) StoreL2BridgeMessages(messages []L2BridgeMessage) error {
	result := db.gorm.Create(&messages)
	return result.Error
}

func (db bridgeMessagesDB) L2BridgeMessage(msgHash common.Hash) (*L2BridgeMessage, error) {
	return db.L2BridgeMessageWithFilter(BridgeMessage{MessageHash: msgHash})
}

func (db bridgeMessagesDB) L2BridgeMessageWithFilter(filter BridgeMessage) (*L2BridgeMessage, error) {
	var sentMessage L2BridgeMessage
	result := db.gorm.Where(&filter).Take(&sentMessage)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &sentMessage, nil
}

func (db bridgeMessagesDB) MarkRelayedL2BridgeMessage(messageHash common.Hash, relayEvent uuid.UUID) error {
	message, err := db.L2BridgeMessage(messageHash)
	if err != nil {
		return err
	} else if message == nil {
		return fmt.Errorf("L2BridgeMessage with message hash %s not found", messageHash)
	}

	message.RelayedMessageEventGUID = &relayEvent
	result := db.gorm.Save(message)
	return result.Error
}
