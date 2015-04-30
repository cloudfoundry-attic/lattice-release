package receptor

import "github.com/tedsuo/rata"

const (
	// Tasks
	CreateTaskRoute = "CreateTask"
	TasksRoute      = "Tasks"
	GetTaskRoute    = "GetTask"
	DeleteTaskRoute = "DeleteTask"
	CancelTaskRoute = "CancelTask"

	// DesiredLRPs
	CreateDesiredLRPRoute = "CreateDesiredLRP"
	GetDesiredLRPRoute    = "GetDesiredLRP"
	UpdateDesiredLRPRoute = "UpdateDesiredLRP"
	DeleteDesiredLRPRoute = "DeleteDesiredLRP"
	DesiredLRPsRoute      = "DesiredLRPs"

	// ActualLRPs
	ActualLRPsRoute                         = "ActualLRPs"
	ActualLRPsByProcessGuidRoute            = "ActualLRPsByProcessGuid"
	ActualLRPByProcessGuidAndIndexRoute     = "ActualLRPByProcessGuidAndIndex"
	KillActualLRPByProcessGuidAndIndexRoute = "KillActualLRPByProcessGuidAndIndex"

	// Cells
	CellsRoute = "Cells"

	// Domains
	UpsertDomainRoute = "UpsertDomain"
	DomainsRoute      = "Domains"

	// Event Streaming
	EventStream = "EventStream"

	// Authentication Cookie
	GenerateCookie = "GenerateCookie"
)

var Routes = rata.Routes{
	// Tasks
	{Path: "/v1/tasks", Method: "POST", Name: CreateTaskRoute},
	{Path: "/v1/tasks", Method: "GET", Name: TasksRoute},
	{Path: "/v1/tasks/:task_guid", Method: "GET", Name: GetTaskRoute},
	{Path: "/v1/tasks/:task_guid", Method: "DELETE", Name: DeleteTaskRoute},
	{Path: "/v1/tasks/:task_guid/cancel", Method: "POST", Name: CancelTaskRoute},

	// DesiredLRPs
	{Path: "/v1/desired_lrps", Method: "GET", Name: DesiredLRPsRoute},
	{Path: "/v1/desired_lrps", Method: "POST", Name: CreateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "GET", Name: GetDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "PUT", Name: UpdateDesiredLRPRoute},
	{Path: "/v1/desired_lrps/:process_guid", Method: "DELETE", Name: DeleteDesiredLRPRoute},

	// ActualLRPs
	{Path: "/v1/actual_lrps", Method: "GET", Name: ActualLRPsRoute},
	{Path: "/v1/actual_lrps/:process_guid", Method: "GET", Name: ActualLRPsByProcessGuidRoute},
	{Path: "/v1/actual_lrps/:process_guid/index/:index", Method: "GET", Name: ActualLRPByProcessGuidAndIndexRoute},
	{Path: "/v1/actual_lrps/:process_guid/index/:index", Method: "DELETE", Name: KillActualLRPByProcessGuidAndIndexRoute},

	// Cells
	{Path: "/v1/cells", Method: "GET", Name: CellsRoute},

	// Domains
	{Path: "/v1/domains/:domain", Method: "PUT", Name: UpsertDomainRoute},
	{Path: "/v1/domains", Method: "GET", Name: DomainsRoute},

	// Event Streaming
	{Path: "/v1/events", Method: "GET", Name: EventStream},

	// Authentication Cookie
	{Path: "/v1/auth_cookie", Method: "POST", Name: GenerateCookie},
}
