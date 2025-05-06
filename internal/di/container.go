// Package di provides dependency injection facilities for the application
package di

import (
	"mxclone/adapters/primary"
	"mxclone/adapters/secondary"
	"mxclone/ports/input"
)

// Container represents a simple dependency injection container
type Container struct {
	// Services (primary adapters implementing input ports)
	dnsService          input.DNSPort
	dnsblService        input.DNSBLPort
	smtpService         input.SMTPPort
	emailAuthService    input.EmailAuthPort
	networkToolsService input.NetworkToolsPort
}

// NewContainer creates a new dependency injection container with all services properly wired up
func NewContainer() *Container {
	// Create repositories (secondary adapters implementing output ports)
	dnsRepository := secondary.NewDNSRepository()

	// Create core services and wire up dependencies
	dnsService := primary.NewDNSAdapter(dnsRepository)

	dnsblRepository := secondary.NewDNSBLRepository(dnsService)
	dnsblService := primary.NewDNSBLAdapter(dnsblRepository)

	smtpRepository := secondary.NewSMTPRepository(dnsService)
	smtpService := primary.NewSMTPAdapter(smtpRepository)

	emailAuthRepository := secondary.NewEmailAuthRepository(dnsService)
	emailAuthService := primary.NewEmailAuthAdapter(emailAuthRepository)

	networkToolsRepository := secondary.NewNetworkToolsRepository()
	networkToolsService := primary.NewNetworkToolsAdapter(networkToolsRepository)

	return &Container{
		dnsService:          dnsService,
		dnsblService:        dnsblService,
		smtpService:         smtpService,
		emailAuthService:    emailAuthService,
		networkToolsService: networkToolsService,
	}
}

// GetDNSService returns the DNS service
func (c *Container) GetDNSService() input.DNSPort {
	return c.dnsService
}

// GetDNSBLService returns the DNSBL service
func (c *Container) GetDNSBLService() input.DNSBLPort {
	return c.dnsblService
}

// GetSMTPService returns the SMTP service
func (c *Container) GetSMTPService() input.SMTPPort {
	return c.smtpService
}

// GetEmailAuthService returns the Email Authentication service
func (c *Container) GetEmailAuthService() input.EmailAuthPort {
	return c.emailAuthService
}

// GetNetworkToolsService returns the Network Tools service
func (c *Container) GetNetworkToolsService() input.NetworkToolsPort {
	return c.networkToolsService
}
