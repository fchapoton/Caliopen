// Copyleft (ɔ) 2017 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package Notifications

import (
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.backends"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.backends/index/elasticsearch"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.backends/store/cassandra"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/nats-io/go-nats"
	"os"
)

type (
	Notifiers interface {
		ByEmail(*Notification) error
		ByNotifQueue(*Notification) error
		BySMS(*Notification) error
		ByIRC(*Notification) error
		ByXMPP(*Notification) error
		ByMobilePhonePush(*Notification) error
	}

	Notifier struct {
		admin        *User          // Admin user on whose behalf actions could be done
		adminLocalID *LocalIdentity // Admin's local identity used to send emails
		config       *NotifierConfig
		index        backends.NotificationsIndex
		natsQueue    *nats.Conn
		natsTopics   map[string]string
		store        backends.NotificationsStore
		log          *log.Logger
	}
)

// NewNotificationsFacility initialises the notifiers
// it takes the same store & index configurations than the REST API for now
func NewNotificationsFacility(config CaliopenConfig, queue *nats.Conn) (notifier *Notifier) {
	notifier = new(Notifier)
	notifier.log = log.New()
	notifier.log.Out = os.Stdout
	// It could set this to any `io.Writer` such as a file
	// file, err := os.OpenFile("notifications.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	//  notifier.log.Out = file
	// } else {
	//  log.Info("Failed to log to file, using default stdout")
	//  notifier.log.Out = os.Stdout
	// }
	notifier.natsQueue = queue
	notifier.config = &config.NotifierConfig
	notifier.natsTopics = make(map[string]string)
	notifier.natsTopics[Nats_outSMTP_topicKey] = config.NatsConfig.OutSMTP_topic
	notifier.natsTopics[Nats_Contacts_topicKey] = config.NatsConfig.Contacts_topic
	switch config.RESTstoreConfig.BackendName {
	case "cassandra":
		cassaConfig := store.CassandraConfig{
			Hosts:       config.RESTstoreConfig.Hosts,
			Keyspace:    config.RESTstoreConfig.Keyspace,
			Consistency: gocql.Consistency(config.RESTstoreConfig.Consistency),
		}
		if config.RESTstoreConfig.ObjStoreType == "s3" {
			cassaConfig.WithObjStore = true
			cassaConfig.OSSConfig.Endpoint = config.RESTstoreConfig.OSSConfig.Endpoint
			cassaConfig.OSSConfig.AccessKey = config.RESTstoreConfig.OSSConfig.AccessKey
			cassaConfig.OSSConfig.SecretKey = config.RESTstoreConfig.OSSConfig.SecretKey
			cassaConfig.OSSConfig.Location = config.RESTstoreConfig.OSSConfig.Location
			cassaConfig.OSSConfig.RawMsgBucket = config.RESTstoreConfig.OSSConfig.Buckets["raw_messages"]
			cassaConfig.OSSConfig.AttachmentBucket = config.RESTstoreConfig.OSSConfig.Buckets["temporary_attachments"]
		}
		backend, err := store.InitializeCassandraBackend(cassaConfig)
		if err != nil {
			log.WithError(err).Fatalf("Initalization of %s backend failed", config.RESTstoreConfig.BackendName)
		}
		notifier.store = backends.NotificationsStore(backend) // type conversion
	default:
		log.Fatalf("Unknown backend: %s", config.RESTstoreConfig.BackendName)
	}

	switch config.RESTindexConfig.IndexName {
	case "elasticsearch":
		esConfig := index.ElasticSearchConfig{
			Urls: config.RESTindexConfig.Hosts,
		}
		index, err := index.InitializeElasticSearchIndex(esConfig)
		if err != nil {
			log.WithError(err).Fatalf("Initalization of %s index failed", config.RESTindexConfig.IndexName)
		}
		notifier.index = backends.NotificationsIndex(index) // type conversion
	default:
		log.Fatalf("Unknown index: %s", config.RESTindexConfig.IndexName)
	}

	user, err := notifier.store.UserByUsername(config.NotifierConfig.AdminUsername)
	if err != nil {
		log.WithError(err).Warnf("Failed to retrieve admin user <%s>", config.NotifierConfig.AdminUsername)
	} else if user != nil {
		notifier.admin = user
		ids, err := notifier.store.GetLocalsIdentities(user.UserId.String())
		if err != nil {
			log.WithError(err).Warnf("Failed to retrieve local identities for admin user <%s>", config.NotifierConfig.AdminUsername)
		} else {
			// get first local identity found for now
			notifier.adminLocalID = &(ids[0])
		}
	}

	return notifier
}

func (N *Notifier) LogNotification(notif *Notification) {
	N.log.WithFields(log.Fields{
		"method": "ByEmail",
		"id":     notif.Id.String(),
	}).Infof("[Notifier] a notification has been issued for user %s", notif.User.UserId.String())
}
