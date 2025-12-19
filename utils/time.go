package utils
import ( 
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func ParseDate(s string) time.Time {
	layout := "2006-01-02"
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}