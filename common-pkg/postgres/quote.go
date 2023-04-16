package postgres

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func QuoteArg(arg any) string {
	var quote string

	switch arg := arg.(type) {
	case nil:
		quote = "null"
	case int64:
		quote = strconv.FormatInt(arg, 10)
	case float64:
		quote = strconv.FormatFloat(arg, 'f', -1, 64)
	case bool:
		quote = strconv.FormatBool(arg)
	case time.Time:
		quote = arg.Format("'2006-01-02 15:04:05.999999999Z07:00:00'")
	case []byte:
		quote = fmt.Sprintf(`'\x%s'::bytea`, hex.EncodeToString(arg))
	case string:
		quote = fmt.Sprintf(`'%s'`, strings.Replace(arg, "'", "''", -1))
	}

	return quote
}
