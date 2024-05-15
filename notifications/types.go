package notifications

type Notification interface {
	SetUp() error
	Send(data []byte, emoji string) error
}

// Ntfy is a struct that contains the configuration for the Ntfy notification service.
// Visit https://docs.ntfy.sh/ for more information.
type Ntfy struct {
	// AccessToken is the access token for the Ntfy service. Required if the Ntfy service requires authentication using access token. Provide the access token in the environment variable NOTIFICATION_NTFY_AUTH_TOKEN.
	AccessToken string `yaml:"-"`
	// ServerEndpoint is the endpoint for the Ntfy service.
	ServerEndpoint string `yaml:"serverEndpoint"`
	// SubscriptionID is the subscription ID for the Ntfy service.
	SubscriptionID string `yaml:"subscriptionID"`
	// UserName is the username for the Ntfy service. Required if the Ntfy service requires authentication using username and password. Provide the password in the environment variable NOTIFICATION_NTFY_PASSWORD.
	UserName string `yaml:"userName"`
	// Password is the password for the Ntfy service. Required if the Ntfy service requires authentication using username and password. Provide the password in the environment variable NOTIFICATION_NTFY_PASSWORD.
	Password string `yaml:"-"`
}
