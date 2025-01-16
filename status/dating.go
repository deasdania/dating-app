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

	UserErrCode_Generic                         DatingStatusCode = "DAPPSXXX4000"
	UserErrCode_Unauthorized                    DatingStatusCode = "DAPPSXXX4001"
	UserErrCode_MissingCredentials              DatingStatusCode = "DAPPSXXX4002"
	UserErrCode_InvalidRequest                  DatingStatusCode = "DAPPSXXX4003"
	UserErrCode_InvalidRequestProfileIDRequired DatingStatusCode = "DAPPSXXX4004"
	UserErrCode_InvalidRequestDirectionRequired DatingStatusCode = "DAPPSXXX4005"
	UserErrCode_NotFoundDating                  DatingStatusCode = "DAPPSXXX4006"
	UserErrCode_ProfileNotFound                 DatingStatusCode = "DAPPSXXX4007"
	UserErrCode_ReachDailyLimit                 DatingStatusCode = "DAPPSXXX4008"
	UserErrCode_AlreadySwiped                   DatingStatusCode = "DAPPSXXX4009"

	SystemErrCode_Generic                       DatingStatusCode = "DAPPSXXX5000"
	SystemErrCode_FailedReadMetadata            DatingStatusCode = "DAPPSXXX5002"
	SystemErrCode_FailedReadOrgID               DatingStatusCode = "DAPPSXXX5003"
	SystemErrCode_FailedSanitize                DatingStatusCode = "DAPPSXXX5004"
	SystemErrCode_FailedStoreData               DatingStatusCode = "DAPPSXXX5005"
	SystemErrCode_FailedBrowseData              DatingStatusCode = "DAPPSXXX5006"
	SystemErrCode_FailedStartTransaction        DatingStatusCode = "DAPPSXXX5008"
	SystemErrCode_FailedEndTransaction          DatingStatusCode = "DAPPSXXX5009"
	SystemErrCode_FailedCommitTransaction       DatingStatusCode = "DAPPSXXX5010"
	SystemErrCode_FailedSwipeTracking           DatingStatusCode = "DAPPSXXX5011"
	SystemErrCode_FailedSwipeCount              DatingStatusCode = "DAPPSXXX5012"
	SystemErrCode_FailedParseSwipe              DatingStatusCode = "DAPPSXXX5013"
	SystemErrCode_FailedSwipeAddingProfile      DatingStatusCode = "DAPPSXXX5014"
	SystemErrCode_FailedSwipeUpdatingSwipeCount DatingStatusCode = "DAPPSXXX5015"
	SystemErrCode_FailedSwipeSettingExpire      DatingStatusCode = "DAPPSXXX5016"
)

var datingMap = map[DatingStatusCode]StatusResponse{
	// 2XXX - success
	Success_Generic:    {StatusDesc: "completed"},
	Success_Processing: {StatusDesc: "retrieving dating"},

	// 4XXX - user errors
	UserErrCode_Generic:                         {StatusDesc: "generic user error"},
	UserErrCode_Unauthorized:                    {StatusDesc: "Unauthorized"},
	UserErrCode_MissingCredentials:              {StatusDesc: "missing credentials"},
	UserErrCode_InvalidRequest:                  {StatusDesc: "Invalid request: %s"},
	UserErrCode_InvalidRequestProfileIDRequired: {StatusDesc: "profile_id is required (uuid)"},
	UserErrCode_InvalidRequestDirectionRequired: {StatusDesc: "direction is required ('left' or 'right')"},
	UserErrCode_NotFoundDating:                  {StatusDesc: "not found dating"},
	UserErrCode_ProfileNotFound:                 {StatusDesc: "profile not found"},
	UserErrCode_ReachDailyLimit:                 {StatusDesc: "you have reached your daily swipe limit."},

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

	SystemErrCode_FailedSwipeTracking: {StatusDesc: "error initializing swipe tracking: %v"},
	SystemErrCode_FailedSwipeCount:    {StatusDesc: "error counting swipes: %v"},
	SystemErrCode_FailedParseSwipe:    {StatusDesc: "error parsing swipes count: %v"},

	SystemErrCode_FailedSwipeAddingProfile:      {StatusDesc: "error adding profile to swiped set: %v"},
	SystemErrCode_FailedSwipeUpdatingSwipeCount: {StatusDesc: "error updating swipe count: %v"},
	SystemErrCode_FailedSwipeSettingExpire:      {StatusDesc: "error setting expiration for swipes key: %v"},
}
