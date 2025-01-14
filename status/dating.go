package status

// Please add the error code sequentially between the status code and status desc and double check the sequence

type StatusResponse struct {
	// status_code contains the status code of the request.
	StatusCode string `json:"status_code,omitempty"`
	// status_desc contains the accompanying description of the given status code of the request.
	StatusDesc string `json:"status_desc,omitempty"`
}

const (
	Success_Generic    DatingStatusCode = "DAPPSXXX2000"
	Success_Processing DatingStatusCode = "DAPPSXXX2001"

	UserErrCode_Generic            DatingStatusCode = "DAPPSXXX4000"
	UserErrCode_Unauthorized       DatingStatusCode = "DAPPSXXX4001"
	UserErrCode_MissingCredentials DatingStatusCode = "DAPPSXXX4002"
	UserErrCode_InvalidRequest     DatingStatusCode = "DAPPSXXX4003"
	UserErrCode_NotFoundDating     DatingStatusCode = "DAPPSXXX4004"

	SystemErrCode_Generic                 DatingStatusCode = "DAPPSXXX5000"
	SystemErrCode_FailedReadMetadata      DatingStatusCode = "DAPPSXXX5002"
	SystemErrCode_FailedReadOrgID         DatingStatusCode = "DAPPSXXX5003"
	SystemErrCode_FailedSanitize          DatingStatusCode = "DAPPSXXX5004"
	SystemErrCode_FailedStoreData         DatingStatusCode = "DAPPSXXX5005"
	SystemErrCode_FailedBrowseData        DatingStatusCode = "DAPPSXXX5006"
	SystemErrCode_FailedStartTransaction  DatingStatusCode = "DAPPSXXX5008"
	SystemErrCode_FailedEndTransaction    DatingStatusCode = "DAPPSXXX5009"
	SystemErrCode_FailedCommitTransaction DatingStatusCode = "DAPPSXXX5010"
)

var datingMap = map[DatingStatusCode]StatusResponse{
	// 2XXX - success
	Success_Generic:    {StatusDesc: "completed"},
	Success_Processing: {StatusDesc: "retrieving dating"},

	// 4XXX - user errors
	UserErrCode_Generic:            {StatusDesc: "generic user error"},
	UserErrCode_Unauthorized:       {StatusDesc: "Unauthorized"},
	UserErrCode_MissingCredentials: {StatusDesc: "missing credentials"},
	UserErrCode_InvalidRequest:     {StatusDesc: "Invalid request: %s"},
	UserErrCode_NotFoundDating:     {StatusDesc: "not found dating"},

	// 5XXX - system errors
	SystemErrCode_Generic:                 {StatusDesc: "generic system error"},
	SystemErrCode_FailedReadMetadata:      {StatusDesc: "could not read metadata from context"},
	SystemErrCode_FailedReadOrgID:         {StatusDesc: "could not read organization ID from metadata"},
	SystemErrCode_FailedSanitize:          {StatusDesc: "failed to sanitize parameters"},
	SystemErrCode_FailedStoreData:         {StatusDesc: "failed to store data"},
	SystemErrCode_FailedBrowseData:        {StatusDesc: "failed to browse data"},
	SystemErrCode_FailedStartTransaction:  {StatusDesc: "failed to start data the transaction"},
	SystemErrCode_FailedEndTransaction:    {StatusDesc: "failed to end data the transaction"},
	SystemErrCode_FailedCommitTransaction: {StatusDesc: "failed to commit the transaction"},
}
