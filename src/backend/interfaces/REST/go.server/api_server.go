// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package rest_api

import (
	obj "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/middlewares"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/contacts"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/devices"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/discussions"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/identities"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/imports"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/messages"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/notifications"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/participants"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/providers"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/tags"
	"github.com/CaliOpen/Caliopen/src/backend/interfaces/REST/go.server/operations/users"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.main"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/loads"
	"os"
)

var (
	server *REST_API
)

type (
	REST_API struct {
		config   APIConfig
		swagSpec *loads.Document
	}

	APIConfig struct {
		Interface      string `mapstructure:"listen_interface"`
		ListenPort     string `mapstructure:"listen_port"`
		Port           string `mapstructure:"port"`
		Hostname       string `mapstructure:"hostname"`
		SwaggerFile    string `mapstructure:"swaggerSpec"`
		BackendConfig  `mapstructure:"BackendConfig"`
		IndexConfig    `mapstructure:"IndexConfig"`
		CacheSettings  `mapstructure:"RedisConfig"`
		NatsConfig     `mapstructure:"NatsConfig"`
		NotifierConfig `mapstructure:"NotifierConfig"`
		Providers      []obj.Provider `mapstructure:"Providers"`
	}

	BackendConfig struct {
		BackendName string          `mapstructure:"backend_name"`
		Settings    BackendSettings `mapstructure:"backend_settings"`
	}

	BackendSettings struct {
		Hosts            []string        `mapstructure:"hosts"`
		Keyspace         string          `mapstructure:"keyspace"`
		Consistency      uint16          `mapstructure:"consistency_level"`
		SizeLimit        uint64          `mapstructure:"raw_size_limit"` // max size for db (in bytes)
		ObjStoreType     string          `mapstructure:"object_store"`
		ObjStoreSettings obj.OSSConfig   `mapstructure:"object_store_settings"`
		UseVault         bool            `mapstructure:"use_vault"`
		VaultSettings    obj.VaultConfig `mapstructure:"vault_settings"`
	}

	IndexConfig struct {
		IndexName string        `mapstructure:"index_name"`
		Settings  IndexSettings `mapstructure:"index_settings"`
	}

	IndexSettings struct {
		Hosts []string `mapstructure:"hosts"`
	}

	CacheSettings struct {
		Host     string `mapstructure:"host"`
		Password string `mapstructure:"password"`
		Db       int    `mapstructure:"db"`
	}

	NatsConfig struct {
		Url              string `mapstructure:"url"`
		NatsQueue        string `mapstructure:"nats_queue"`
		OutSMTP_topic    string `mapstructure:"outSMTP_topic"`
		OutIMAP_topic    string `mapstructure:"outIMAP_topic"`
		OutTWITTER_topic string `mapstructure:"outTWITTER_topic"`
		Contacts_topic   string `mapstructure:"contacts_topic"`
		Keys_topic       string `mapstructure:"keys_topic"`
		Users_topic      string `mapstructure:"users_topic"`
		IdPoller_topic   string `mapstructure:"idpoller_topic"`
	}

	NotifierConfig struct {
		AdminUsername string `mapstructure:"admin_username"`
		BaseUrl       string `mapstructure:"base_url"`
		TemplatesPath string `mapstructure:"templates_path"`
	}
)

func InitializeServer(config APIConfig) error {
	server = new(REST_API)
	return server.initialize(config)
}

func (server *REST_API) initialize(config APIConfig) error {
	server.config = config

	//init Caliopen facility
	caliopenConfig := obj.CaliopenConfig{
		RESTstoreConfig: obj.RESTstoreConfig{
			BackendName:  config.BackendName,
			Hosts:        config.BackendConfig.Settings.Hosts,
			Keyspace:     config.BackendConfig.Settings.Keyspace,
			Consistency:  config.BackendConfig.Settings.Consistency,
			SizeLimit:    config.BackendConfig.Settings.SizeLimit,
			ObjStoreType: config.BackendConfig.Settings.ObjStoreType,
			OSSConfig:    config.BackendConfig.Settings.ObjStoreSettings,
			UseVault:     config.BackendConfig.Settings.UseVault,
			VaultConfig:  config.BackendConfig.Settings.VaultSettings,
		},
		RESTindexConfig: obj.RESTIndexConfig{
			IndexName: config.IndexConfig.IndexName,
			Hosts:     config.IndexConfig.Settings.Hosts,
		},
		CacheConfig: obj.CacheConfig{
			Host:     config.CacheSettings.Host,
			Password: config.CacheSettings.Password,
			Db:       config.CacheSettings.Db,
		},
		NatsConfig: obj.NatsConfig{
			Url:              config.NatsConfig.Url,
			NatsQueue:        config.NatsConfig.NatsQueue,
			OutSMTP_topic:    config.NatsConfig.OutSMTP_topic,
			OutIMAP_topic:    config.NatsConfig.OutIMAP_topic,
			OutTWITTER_topic: config.NatsConfig.OutTWITTER_topic,
			Contacts_topic:   config.NatsConfig.Contacts_topic,
			Keys_topic:       config.NatsConfig.Keys_topic,
			Users_topic:      config.NatsConfig.Users_topic,
			IdPoller_topic:   config.NatsConfig.IdPoller_topic,
		},
		NotifierConfig: obj.NotifierConfig{
			AdminUsername: config.NotifierConfig.AdminUsername,
			BaseUrl:       config.NotifierConfig.BaseUrl,
			TemplatesPath: config.NotifierConfig.TemplatesPath,
		},
		Providers: config.Providers,
		Hostname:  config.Hostname + ":" + config.Port,
	}

	err := caliopen.Initialize(caliopenConfig)

	if err != nil {
		log.WithError(err).Fatal("Caliopen facilities initialization failed")
	}

	//checks that with could open the swagger specs file
	_, err = os.Stat(server.config.SwaggerFile)
	if err != nil {
		return err
	}

	return nil
}

func StartServer() error {
	return server.start()
}

func (server *REST_API) start() error {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	//router.Use(Dumper())

	// adds our middlewares
	err := http_middleware.InitSwaggerMiddleware(server.config.SwaggerFile)
	if err != nil {
		log.WithError(err).Warn("init swagger middleware failed")
	} else {
		router.Use(http_middleware.SwaggerValidator())
	}
	// adds our routes and handlers
	api := router.Group(http_middleware.RoutePrefix)
	server.AddHandlers(api)

	// listens
	addr := server.config.Interface + ":" + server.config.ListenPort
	err = router.Run(addr)
	if err != nil {
		log.WithError(err).Warn("unable to start gin server")
	}
	return err
}

func (server *REST_API) AddHandlers(api *gin.RouterGroup) {

	/** users API **/
	usrs := api.Group("/users", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	usrs.PATCH("/:user_id", users.PatchUser)
	usrs.POST("/:user_id/actions", users.Delete)

	/** identities **/
	ids := api.Group(http_middleware.IdentitiesRoute, http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	ids.GET("/locals", identities.GetLocalsIdentities)
	ids.GET("/locals/:identity_id", identities.GetLocalIdentity)
	ids.GET("/remotes", identities.GetRemoteIdentities)
	ids.POST("/remotes", identities.NewRemoteIdentity)
	ids.GET("/remotes/:remote_id", identities.GetRemoteIdentity)
	ids.PATCH("/remotes/:remote_id", identities.PatchRemoteIdentity)
	ids.DELETE("/remotes/:remote_id", identities.DeleteRemoteIdentity)

	/** passwords API **/
	passwords := api.Group("/passwords")
	passwords.GET("/reset", notImplemented)
	passwords.POST("/reset", users.RequestPasswordReset)
	passwords.GET("/reset/:reset_token", users.ValidatePassResetToken)
	passwords.POST("/reset/:reset_token", users.ResetPassword)

	/** username API **/
	api.GET("/username/isAvailable", users.IsAvailable)

	/** messages API **/
	msg := api.Group("/messages", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	msg.GET("", messages.GetMessagesList)
	msg.GET("/:message_id", messages.GetMessage)
	msg.POST("/:message_id/actions", messages.Actions)
	//attachments
	msg.POST("/:message_id/attachments", messages.UploadAttachment)
	msg.DELETE("/:message_id/attachments/:attachment_id", messages.DeleteAttachment)
	msg.GET("/:message_id/attachments/:attachment_id", messages.DownloadAttachment)
	//tags
	msg.PATCH("/:message_id/tags", tags.PatchResourceWithTags)

	/** discussions API **/
	disc := api.Group("/discussions", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	disc.GET("", discussions.GetDiscussionsList)
	disc.GET("/:discussionId", discussions.GetDiscussion)

	/** participants API **/
	parts := api.Group("/participants", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	parts.GET("/suggest", participants.Suggest)
	parts.POST("/discussion", participants.HashUris)

	/** contacts API **/
	cts := api.Group(http_middleware.ContactsRoute, http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	cts.GET("", contacts.GetContactsList)
	cts.POST("", contacts.NewContact)
	cts.GET("/:contactID", contacts.GetContact)
	cts.PATCH("/:contactID", contacts.PatchContact)
	cts.DELETE("/:contactID", contacts.DeleteContact)
	cts.GET("/:contactID/identities", contacts.GetIdentities)
	//publickeys
	cts.POST("/:contactID/publickeys", contacts.NewPublicKey)
	cts.GET("/:contactID/publickeys", contacts.GetPubKeys)
	cts.GET("/:contactID/publickeys/:pubkeyID", contacts.GetPubKey)
	cts.PATCH("/:contactID/publickeys/:pubkeyID", contacts.PatchPubKey)
	cts.DELETE("/:contactID/publickeys/:pubkeyID", contacts.DeletePubKey)
	//tags
	cts.PATCH("/:contactID/tags", tags.PatchResourceWithTags)

	/** devices API **/
	api.GET("/validate-device/:token", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"), devices.ValidateDevice)
	dev := api.Group(http_middleware.DevicesRoute, http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	dev.GET("", devices.GetDevicesList)
	//dev.POST("", devices.NewDevice)
	dev.GET("/:deviceID", devices.GetDevice)
	dev.PATCH("/:deviceID", devices.PatchDevice)
	dev.DELETE("/:deviceID", devices.DeleteDevice)
	dev.POST("/:deviceID/actions", devices.Actions)

	// imports
	imp := api.Group(http_middleware.ImportsRoute, http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	imp.POST("", imports.ImportFile)

	/** tags API **/
	tag := api.Group(http_middleware.TagsRoute, http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	tag.GET("", tags.RetrieveUserTags)
	tag.POST("", tags.CreateTag)
	tag.GET("/:tag_name", tags.RetrieveTag)
	tag.PATCH("/:tag_name", tags.PatchTag)
	tag.DELETE("/:tag_name", tags.DeleteTag)

	/** search API **/
	search := api.Group("/search", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	search.GET("", operations.SimpleSearch)
	search.POST("", operations.AdvancedSearch)

	/** notifications API **/
	notif := api.Group("/notifications", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"))
	notif.GET("", notifications.GetPendingNotif)
	notif.DELETE("", notifications.DeleteNotifications)
	notif.GET("/:notification_id", notifications.GetNotification)
	notif.DELETE("/:notification_id", notifications.DeleteNotification)

	/** providers **/
	prov := api.Group("/providers")
	prov.GET("", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"), providers.GetProvidersList)
	prov.GET("/:provider_name", http_middleware.BasicAuthFromCache(caliopen.Facilities.Cache, "caliopen"), providers.GetProvider)
	prov.GET("/:provider_name/callback", providers.CallbackHandler)
	api.StaticFile("/test/oauth", "../interfaces/REST/go.server/operations/providers/oauth-test.html")
}
