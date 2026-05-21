package nendesktop

import "encoding/json"

// Desktop is the response type for all desktop operations.
type Desktop struct {
	DesktopID     string       `json:"desktop_id"`
	DesktopType   string       `json:"desktop_type"`
	Status        string       `json:"status"`
	WorkspaceID   string       `json:"workspace_id"`
	InstanceID    string       `json:"instance_id,omitempty"`
	PublicIP      string       `json:"public_ip,omitempty"`
	PrivateIP     string       `json:"private_ip,omitempty"`
	Name          string       `json:"name,omitempty"`
	ControllerArn string       `json:"controller_arn,omitempty"`
	Session       *SessionInfo `json:"session,omitempty"`
	CreatedAt     int64        `json:"created_at"`
	UpdatedAt     int64        `json:"updated_at"`
}

// SessionInfo is the RDP session sub-resource.
type SessionInfo struct {
	Href        string `json:"href"`
	Active      bool   `json:"active"`
	Interactive bool   `json:"interactive"`
}

// DeleteResponse is returned by DeleteDesktop.
type DeleteResponse struct {
	DesktopID string `json:"desktop_id"`
	Status    string `json:"status"`
}

// ExecuteResult is the response from POST /desktops/:id/execute.
type ExecuteResult struct {
	Status      string `json:"status"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	Base64Image string `json:"base64_image,omitempty"`
}

// ToolSchema describes a single tool exposed by a desktop.
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// executeRequest is the internal request body for POST /desktops/:id/execute.
type executeRequest struct {
	Action executeAction `json:"action"`
}

type executeAction struct {
	Tool   string         `json:"tool"`
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
}

// createDesktopRequest is the internal request body for POST /desktops.
type createDesktopRequest struct {
	DesktopType string `json:"desktop_type"`
}

// updateDesktopRequest is the internal request body for PATCH /desktops/:id.
type updateDesktopRequest struct {
	Name string `json:"name"`
}

// toolLogsResponse is used to decode the opaque tool-logs JSON array.
type toolLogsResponse = json.RawMessage

// File describes a single entry returned by ListFiles. Modified is a
// Unix-epoch float with fractional seconds. IsDir is true for
// directory entries; the field is omitted from the wire when false so
// older servers that don't emit it keep deserializing unchanged.
type File struct {
	Name     string  `json:"name"`
	Size     int64   `json:"size"`
	Modified float64 `json:"modified"`
	IsDir    bool    `json:"is_dir,omitempty"`
}

// UploadFileResponse is returned by UploadFile (POST /desktops/:id/files/:name).
type UploadFileResponse struct {
	Success  bool   `json:"success"`
	Size     int64  `json:"size"`
	Filename string `json:"filename"`
}
