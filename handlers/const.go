package handlers

const (
	success = "success"
	failure = "failure"
)

// func parseNotUserID(idEnv string) int64 {
// 	idStr := os.Getenv(idEnv)
// 	if len(idStr) >= 3 && idStr[:3] != "-100" {
// 		idStr = "-100" + idStr
// 	}

// 	id, err := strconv.ParseInt(idStr, 10, 64)
// 	if err != nil {
// 		log.WithField(logconst.FieldFunc, "parseNotUserID").
// 			WithField(logconst.FieldConst, idEnv).
// 			Fatal(err)
// 	}
// 	return id
// }
