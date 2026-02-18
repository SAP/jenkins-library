package btp

import "time"

type RunSyncOptions struct {
	CmdScript      []string
	TimeoutSeconds int
	PollInterval   int
	CheckFunc      func() bool // Function to check the command status
}

type ServiceInstanceData struct {
	ID            string `json:"id"`
	Ready         bool   `json:"ready"`
	LastOperation struct {
		ID                  string    `json:"id"`
		Ready               bool      `json:"ready"`
		Description         string    `json:"description"`
		Type                string    `json:"type"`
		State               string    `json:"state"`
		ResourceID          string    `json:"resource_id"`
		ResourceType        string    `json:"resource_type"`
		PlatformID          string    `json:"platform_id"`
		CorrelationID       string    `json:"correlation_id"`
		Reschedule          bool      `json:"reschedule"`
		RescheduleTimestamp time.Time `json:"reschedule_timestamp"`
		DeletionScheduled   time.Time `json:"deletion_scheduled"`
		CreatedAt           time.Time `json:"created_at"`
		UpdatedAt           time.Time `json:"updated_at"`
	} `json:"last_operation"`
	Name          string `json:"name"`
	ServicePlanID string `json:"service_plan_id"`
	PlatformID    string `json:"platform_id"`
	DashboardURL  string `json:"dashboard_url"`
	Context       struct {
		Origin          string `json:"origin"`
		Region          string `json:"region"`
		ZoneID          string `json:"zone_id"`
		EnvType         string `json:"env_type"`
		Platform        string `json:"platform"`
		Subdomain       string `json:"subdomain"`
		LicenseType     string `json:"license_type"`
		InstanceName    string `json:"instance_name"`
		SubaccountID    string `json:"subaccount_id"`
		CrmCustomerID   string `json:"crm_customer_id"`
		GlobalAccountID string `json:"global_account_id"`
	} `json:"context"`
	Usable       bool      `json:"usable"`
	SubaccountID string    `json:"subaccount_id"`
	Protected    any       `json:"protected"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Labels       string    `json:"labels"`
}

type ServiceBindingData struct {
	ID            string `json:"id"`
	Ready         bool   `json:"ready"`
	LastOperation struct {
		ID                  string    `json:"id"`
		Ready               bool      `json:"ready"`
		Type                string    `json:"type"`
		State               string    `json:"state"`
		ResourceID          string    `json:"resource_id"`
		ResourceType        string    `json:"resource_type"`
		PlatformID          string    `json:"platform_id"`
		CorrelationID       string    `json:"correlation_id"`
		Reschedule          bool      `json:"reschedule"`
		RescheduleTimestamp time.Time `json:"reschedule_timestamp"`
		DeletionScheduled   time.Time `json:"deletion_scheduled"`
		CreatedAt           time.Time `json:"created_at"`
		UpdatedAt           time.Time `json:"updated_at"`
	} `json:"last_operation"`
	Name              string `json:"name"`
	ServiceInstanceID string `json:"service_instance_id"`
	Context           struct {
		Origin            string `json:"origin"`
		Region            string `json:"region"`
		ZoneID            string `json:"zone_id"`
		EnvType           string `json:"env_type"`
		Platform          string `json:"platform"`
		Subdomain         string `json:"subdomain"`
		LicenseType       string `json:"license_type"`
		InstanceName      string `json:"instance_name"`
		SubaccountID      string `json:"subaccount_id"`
		CrmCustomerID     string `json:"crm_customer_id"`
		GlobalAccountID   string `json:"global_account_id"`
		ServiceInstanceID string `json:"service_instance_id"`
	} `json:"context"`
	Credentials struct {
		Abap struct {
			CommunicationArrangementID       string `json:"communication_arrangement_id"`
			CommunicationInboundUserAuthMode string `json:"communication_inbound_user_auth_mode"`
			CommunicationInboundUserID       string `json:"communication_inbound_user_id"`
			CommunicationScenarioID          string `json:"communication_scenario_id"`
			CommunicationSystemID            string `json:"communication_system_id"`
			CommunicationType                string `json:"communication_type"`
			Password                         string `json:"password"`
			Username                         string `json:"username"`
		} `json:"abap"`
		Binding struct {
			Env     string `json:"env"`
			ID      string `json:"id"`
			Type    string `json:"type"`
			Version string `json:"version"`
		} `json:"binding"`
		PreserveHostHeader bool   `json:"preserve_host_header"`
		SapCloudService    string `json:"sap.cloud.service"`
		Systemid           string `json:"systemid"`
		URL                string `json:"url"`
	} `json:"credentials"`
	SubaccountID string    `json:"subaccount_id"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Labels       string    `json:"labels"`
}
