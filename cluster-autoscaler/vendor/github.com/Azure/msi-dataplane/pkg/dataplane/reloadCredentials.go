package dataplane

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
)

type reloadingCredential struct {
	clientOpts   azcore.ClientOptions
	currentValue *azidentity.ClientCertificateCredential
	notBefore    string
	lock         *sync.RWMutex
	logger       *logr.Logger
	ticker       *time.Ticker
}

type Option func(*reloadingCredential)

// WithLogger sets a custom logger for the reloadingCredential.
// This can be useful for debugging or logging purposes.
func WithLogger(logger *logr.Logger) Option {
	return func(c *reloadingCredential) {
		c.logger = logger
	}
}

// WithBackstopRefresh sets a custom timer for the reloadingCredential.
// This can be useful for loading credential file periodically.
func WithBackstopRefresh(d time.Duration) Option {
	return func(c *reloadingCredential) {
		c.ticker = time.NewTicker(d)
	}
}

// WithClientOpts adds common Azure client options. Use this field to, for instance,
// configure the cloud environment in which this credential should authenticate.
func WithClientOpts(o azcore.ClientOptions) Option {
	return func(c *reloadingCredential) {
		c.clientOpts = o
	}
}

// NewUserAssignedIdentityCredential creates a new reloadingCredential for a user-assigned identity.
// ctx is used to manage the lifecycle of the reloader, allowing for cancellation if reloading is no longer needed.
// credentialPath is the path to the credential file.
// opts allows for additional configuration, such as setting a custom logger, periodic reload time, and cloud environment.
//
// The function ensures that a valid token is loaded before returning the credential.
// It also starts a background process to watch for changes to the credential file and reloads it as necessary.
func NewUserAssignedIdentityCredential(ctx context.Context, credentialPath string, opts ...Option) (azcore.TokenCredential, error) {
	defaultLog := logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, nil))
	credential := &reloadingCredential{
		lock:   &sync.RWMutex{},
		logger: &defaultLog,
		ticker: time.NewTicker(6 * time.Hour),
	}

	for _, opt := range opts {
		opt(credential)
	}

	// load once to validate everything and ensure we have a useful token before we return
	if err := credential.load(credentialPath); err != nil {
		return nil, err
	}
	// start the process of watching - the caller can cancel ctx if they want to stop
	if err := credential.start(ctx, credentialPath); err != nil {
		return nil, err
	}
	return credential, nil
}

// GetToken retrieves the current token from the reloadingCredential.
// It uses a read lock to ensure that the token is not being modified while it is being read.
// options specifies additional options for the token request.
func (r *reloadingCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.currentValue.GetToken(ctx, options)
}

func (r *reloadingCredential) start(ctx context.Context, credentialFile string) error {
	// set up the file watcher, call load() when we see events or on some timer in case no events are delivered
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	// we close the file watcher if adding the file to watch fails.
	// this will also close the new go routine created to watch the file
	err = fileWatcher.Add(credentialFile)
	if err != nil {
		if closeErr := fileWatcher.Close(); closeErr != nil {
			r.logger.Error(err, "failed to close file watcher")
		}
		return fmt.Errorf("failed to add credential file to file watcher: %w", err)
	}

	go func() {
		defer func() {
			if err := fileWatcher.Close(); err != nil {
				r.logger.Error(err, "failed to close file watcher")
			}
		}()
		defer r.ticker.Stop()
		for {
			select {
			case event, ok := <-fileWatcher.Events:
				if !ok {
					r.logger.Info("stopping credential reloader since file watcher has no events")
					return
				}
				if event.Op.Has(fsnotify.Write) {
					if err := r.load(credentialFile); err != nil {
						r.logger.Error(err, "failed to reload credential after file event")
					}
				}
			case <-r.ticker.C:
				if err := r.load(credentialFile); err != nil {
					r.logger.Error(err, "failed to reload credential periodically")
				}
			case err, ok := <-fileWatcher.Errors:
				if !ok {
					r.logger.Info("stopping credential reloader since file watcher has no events")
					return
				}
				r.logger.Error(err, "recieved an error from the file watcher")
			case <-ctx.Done():
				r.logger.Info("user signaled context cancel, stopping credential reloader")
				return
			}
		}
	}()
	return nil
}

func (r *reloadingCredential) load(credentialFile string) error {
	// read the file from the filesystem and update the current value we're holding on to if the certificate we read is newer, making sure to not step on the toes of anyone calling GetToken()
	byteValue, err := os.ReadFile(credentialFile)
	if err != nil {
		return fmt.Errorf("failed to read credential file %s: %w", credentialFile, err)
	}

	var credentials UserAssignedIdentityCredentials
	if err := json.Unmarshal(byteValue, &credentials); err != nil {
		return fmt.Errorf("failed to unmarshal credential file %s: %w", credentialFile, err)
	}

	var newCertValue *azidentity.ClientCertificateCredential
	newCertValue, err = GetCredential(r.clientOpts, credentials)
	if err != nil {
		return fmt.Errorf("failed to get client certificate credential: %w", err)
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	if r.notBefore != "" {
		err, ok := isLoadedCredentialNewer(*credentials.NotBefore, r.notBefore)
		if err != nil {
			return fmt.Errorf("failed to determine not_before for credential: %w", err)
		}
		if !ok {
			return nil
		}
	}

	r.currentValue = newCertValue
	r.notBefore = *credentials.NotBefore

	return nil
}

func isLoadedCredentialNewer(newCred string, currentCred string) (error, bool) {
	parsedNewCred, err := time.Parse(time.RFC3339, newCred)
	if err != nil {
		return err, false
	}

	parsedCurrentCred, err := time.Parse(time.RFC3339, currentCred)
	if err != nil {
		return err, false
	}

	return nil, parsedNewCred.After(parsedCurrentCred)
}
