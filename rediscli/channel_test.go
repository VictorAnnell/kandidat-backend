package rediscli

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRedis_ChannelJoinPrivate(t *testing.T) {
	senderUUID := "9999"
	recipientUUID := "9998"

	chMessageX, _, err := testRedisInstance.ChannelJoin(senderUUID, recipientUUID)
	if err != nil {
		log.Fatal(err)
	}

	chMessageY, _, err := testRedisInstance.ChannelJoin(recipientUUID, senderUUID)
	if err != nil {
		log.Fatal(err)
	}

	channelUUIDX, err := testRedisInstance.GetChannelUUID(senderUUID, recipientUUID)
	if err != nil {
		log.Fatal(err)
	}

	channelUUIDY, err := testRedisInstance.GetChannelUUID(recipientUUID, senderUUID)
	if err != nil {
		log.Fatal(err)
	}

	if channelUUIDX != channelUUIDY {
		t.Fatalf("expected channelX [%s] equal channelY [%s]", channelUUIDX, channelUUIDY)
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 100)

			message := &Message{
				UUID:          uuid.NewString(), // fmt.Sprintf("%s%d", id, i+1),
				SenderID:      senderUUID,
				RecipientUUID: recipientUUID,
				Message:       fmt.Sprintf("Helo %s #%d", recipientUUID, i+1),
				CreatedAt:     time.Now(),
			}
			_, err := testRedisInstance.ChannelMessage(message)

			if err != nil {
				t.Error(err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			data := <-chMessageX.Channel()
			log.Println(fmt.Sprintf("X >>> %+v", data))
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			data := <-chMessageY.Channel()
			log.Println(fmt.Sprintf("Y >>> %+v", data))
		}
	}()

	wg.Wait()
}

func TestRedis_ChannelJoinPublic(t *testing.T) {
	senderUUID := "TEST_SENDER"
	recipientUUID := ""

	chMessage, _, err := testRedisInstance.ChannelJoin(senderUUID, recipientUUID)
	if err != nil {
		log.Fatal(err)
	}

	_, err = testRedisInstance.GetChannelUUID(senderUUID, recipientUUID)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 100)

			message := &Message{
				UUID:          uuid.NewString(), // fmt.Sprintf("%s%d", id, i+1),
				SenderID:      senderUUID,
				RecipientUUID: recipientUUID,
				Message:       fmt.Sprintf("Helo %s #%d", recipientUUID, i+1),
				CreatedAt:     time.Now(),
			}

			_, err := testRedisInstance.ChannelMessage(message)
			if err != nil {
				t.Error(err)
			}
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			data := <-chMessage.Channel()
			log.Println(fmt.Sprintf("X >>> %+v", data))
		}
	}()

	wg.Wait()
}
