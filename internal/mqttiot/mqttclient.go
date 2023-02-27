package mqttiot

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	//personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"

	personaliotv1alpha1 "github.com/mgrote/personal-iot/api/v1alpha1"
	"github.com/mgrote/personal-iot/internal"
)

func ClientOpts(mqttConfig personaliotv1alpha1.MQTTConfig) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(*mqttConfig.Broker)
	opts.SetClientID(*mqttConfig.ClientID)
	opts.SetUsername(*mqttConfig.UserName)
	opts.SetPassword(*mqttConfig.Password)
	opts.SetCleanSession(true)
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
