package fort

import (
	"fmt"
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestChaining_FlatMapToJoin(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type TaggedValue struct {
		ID  int
		Tag string
	}

	// 1. FlatMap source into tagged values
	tagged := QueryInformer(&FlatMap[*TaggedValue, int]{
		Lock: lock,
		Map: func(i int) ([]*TaggedValue, error) {
			return []*TaggedValue{
				{ID: i, Tag: "A"},
				{ID: i, Tag: "B"},
			}, nil
		},
		Over: source,
	})

	// 2. Join tagged values with themselves on ID
	joined := QueryInformer(&Join[string, *TaggedValue, *TaggedValue]{
		Lock: lock,
		Select: func(l, r *TaggedValue) (string, error) {
			return fmt.Sprintf("%d:%s-%s", l.ID, l.Tag, r.Tag), nil
		},
		From: tagged,
		Join: tagged,
		On: func(l, r *TaggedValue) any {
			if l != nil {
				return [1]int{l.ID}
			}
			return [1]int{r.ID}
		},
		Where: func(l, r *TaggedValue) bool {
			return l.Tag < r.Tag // Only A-B pairs
		},
	})

	var results []string
	joined.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { results = append(results, obj.(string)) },
	})

	source.OnAdd(1, true)
	// Expected: "1:A-B"
	if len(results) != 1 || results[0] != "1:A-B" {
		t.Errorf("Expected [1:A-B], got %v", results)
	}
}

func TestChaining_JoinToGroupByToFlatMap(t *testing.T) {
	lock := NewLockGroup()
	users := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	orders := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type User struct {
		ID   int
		Name string
	}
	type Order struct {
		ID     int
		UserID int
		Amount int
	}

	// 1. Join Users and Orders
	userOrders := QueryInformer(&Join[UserOrder, *User, *Order]{
		Lock: lock,
		Select: func(u *User, o *Order) (UserOrder, error) {
			return UserOrder{UserName: u.Name, Amount: o.Amount}, nil
		},
		From: users,
		Join: orders,
		On: func(u *User, o *Order) any {
			if u != nil {
				return [1]int{u.ID}
			}
			return [1]int{o.UserID}
		},
	})

	// 2. GroupBy User to get Totals
	type UserTotal struct {
		UserName string
		Total    int64
	}
	userTotals := QueryInformer(&GroupBy[UserTotal, UserOrder]{
		Lock: lock,
		Select: func(fields []GroupField) (UserTotal, error) {
			return UserTotal{
				UserName: fields[0].(string),
				Total:    fields[1].(int64),
			}, nil
		},
		From: userOrders,
		GroupBy: func(uo UserOrder) (any, []GroupField) {
			return [1]string{uo.UserName},
				[]GroupField{
					AnyValue(uo.UserName),
					Sum(int64(uo.Amount)),
				}
		},
	})

	// 3. FlatMap Totals into Alerts (if Total > 100)
	type Alert struct {
		User    string
		Message string
	}
	alerts := QueryInformer(&FlatMap[*Alert, UserTotal]{
		Lock: lock,
		Map: func(ut UserTotal) ([]*Alert, error) {
			if ut.Total > 100 {
				return []*Alert{{User: ut.UserName, Message: "High Spending"}}, nil
			}
			return nil, nil
		},
		Over: userTotals,
	})

	var activeAlerts []*Alert
	alerts.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { activeAlerts = append(activeAlerts, obj.(*Alert)) },
		UpdateFunc: func(old, new any) {
			for i, a := range activeAlerts {
				if a.User == new.(*Alert).User {
					activeAlerts[i] = new.(*Alert)
					return
				}
			}
		},
		DeleteFunc: func(obj any) {
			val := obj.(*Alert)
			for i, a := range activeAlerts {
				if a.User == val.User {
					activeAlerts = append(activeAlerts[:i], activeAlerts[i+1:]...)
					return
				}
			}
		},
	})

	// Pushing data
	users.OnAdd(&User{ID: 1, Name: "Bob"}, true)
	orders.OnAdd(&Order{ID: 101, UserID: 1, Amount: 50}, true)
	if len(activeAlerts) != 0 {
		t.Errorf("Expected 0 alerts for Bob, got %v", activeAlerts)
	}

	// Update Bob's total to 150
	orders.OnAdd(&Order{ID: 102, UserID: 1, Amount: 100}, false)
	if len(activeAlerts) != 1 || activeAlerts[0].User != "Bob" {
		t.Errorf("Expected 1 alert for Bob, got %v", activeAlerts)
	}

	// Reduce Bob's spending (Update order 102)
	orders.OnUpdate(&Order{ID: 102, UserID: 1, Amount: 100}, &Order{ID: 102, UserID: 1, Amount: 10})
	if len(activeAlerts) != 0 {
		t.Errorf("Expected alerts to be cleared for Bob, got %v", activeAlerts)
	}
}
