package api

import (
	"database/sql/driver"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"golang.org/x/crypto/bcrypt"
)

func (ts *Timestamp) Scan(value interface{}) error {
	t, ok := value.(time.Time)

	if ok {
		ts.SetFromGoTime(t)
	}

	return nil
}

func (ts Timestamp) Value() (driver.Value, error) {
	gt := ts.AsTime()

	if gt.IsZero() {
		gt = time.Now()
	}

	return gt, nil
}

func (ts *Timestamp) AsTime() time.Time {
	var t time.Time

	if ts == nil {
		t = time.Unix(0, 0).UTC() // treat nil like the empty Timestamp
	} else {
		t = time.Unix(ts.Seconds, int64(ts.Nanos)).UTC()
	}

	return t
}

func (ts *Timestamp) SetFromGoTime(t time.Time) {
	ts.Seconds = int64(t.Unix())
	ts.Nanos = int32(t.Nanosecond())
}

func (t *TagList) Scan(value interface{}) error {
	v, ok := value.(string)

	if ok {
		if err := jsonpb.UnmarshalString(v, t); err != nil {
			return err
		}
	}

	return nil
}

func (t *TagList) Value() (driver.Value, error) {
	ids := []string{}
	for _, tag := range t.Items {
		ids = append(ids, tag.Id)
	}

	return ids, nil
}

func GeneratePasswordHash(pw string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pw), 0)
}

func CompareHashAndPassword(user *User, pw string) error {
	return bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(pw))
}

func CategoryListInsert(cats []*Category, v *Category, index int) []*Category {
	return append(cats[:index], append([]*Category{v}, cats[index:]...)...)
}

func CategoryListRemove(cats []*Category, index int) []*Category {
	return append(cats[:index], cats[index+1:]...)
}

func CategoryListMove(cats []*Category, srcIndex int, dstIndex int) []*Category {
	v := cats[srcIndex]

	return CategoryListInsert(CategoryListRemove(cats, srcIndex), v, dstIndex)
}
