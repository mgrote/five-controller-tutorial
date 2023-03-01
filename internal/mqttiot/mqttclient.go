package mqttiot

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal"
)

type MQTTMessage struct {
	Topik     string
	Msg       string
	Duplicate bool
}

type MQTTSubscriber interface {
	Connect() error
	Subscribe(topik string, qos byte, messages chan<- MQTTMessage) error
	Unsubscribe(topik string) error
	Disconnect(waitMs uint)
}

type MQTTPublisher interface {
	Connect() error
	Publish(topik string, message string, qos byte, retained bool) error
	Disconnect(waitMs uint)
}

type FakeMQTTPublisher struct {
	ConnectError error
	PublishError error
}

func (f *FakeMQTTPublisher) Publish(topik string, message string, qos byte, retained bool) error {
	return f.PublishError
}

func (f *FakeMQTTPublisher) Connect() error {
	return f.ConnectError
}

func (f *FakeMQTTPublisher) Disconnect(waitMs uint) {
	// nop
}

type FakeMQTTSubscriber struct {
	ConnectError     error
	SubscribeError   error
	UnsubscribeError error
	ExpectedMessages []MQTTMessage
}

func (f *FakeMQTTSubscriber) Connect() error {
	return f.ConnectError
}

func (f *FakeMQTTSubscriber) Subscribe(_ string, _ byte, messages chan<- MQTTMessage) error {
	if f.SubscribeError != nil {
		return f.SubscribeError
	}
	go func() {
		for _, message := range f.ExpectedMessages {
			messages <- message
		}
	}()
	return nil
}

func (f *FakeMQTTSubscriber) Unsubscribe(_ string) error {
	return f.UnsubscribeError
}

func (f *FakeMQTTSubscriber) Disconnect(waitMs uint) {
	// nop
}

func ClientOpts(mqttConfig personaliotv1alpha1.MQTTConfig) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*mqttConfig.Broker)
	opts.SetClientID(*mqttConfig.ClientID)
	opts.SetUsername(*mqttConfig.UserName)
	opts.SetPassword(*mqttConfig.Password)
	opts.SetCleanSession(true)
	opts.SetOrderMatters(true)
	return opts
}

func ClientOptsFromEnv() (*mqtt.ClientOptions, error) {
	opts := mqtt.NewClientOptions()
	broker, found := os.LookupEnv(internal.MqttBroker)
	if !found {
		return nil, fmt.Errorf("unable to find environment var %s", internal.MqttBroker)
	}
	clientID, found := os.LookupEnv(internal.MqttClientID)
	if !found {
		return nil, fmt.Errorf("unable to find environment var %s", internal.MqttClientID)
	}
	user, found := os.LookupEnv(internal.MqttUserName)
	if !found {
		return nil, fmt.Errorf("unable to find environment var %s", internal.MqttUserName)
	}
	pass, found := os.LookupEnv(internal.MqttPassWord)
	if !found {
		return nil, fmt.Errorf("unable to find environment var %s", internal.MqttPassWord)
	}
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.SetCleanSession(true)
	return opts, nil
}
