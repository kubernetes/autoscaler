package fort

import (
	"testing"

	"k8s.io/client-go/tools/cache"
)

type User struct {
	ID   int
	Name string
}

type Order struct {
	ID     int
	UserID int
	Amount int
}

type UserOrder struct {
	UserName string
	Amount   int
}

func TestSelectJoinGroupBy(t *testing.T) {
	lock := NewLockGroup()
	users := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	orders := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	// Query pipeline setup
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
		Where: func(u *User, o *Order) bool {
			return u.ID == o.UserID
		},
	})

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

	var totalResults []UserTotal
	userTotals.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			newT := obj.(UserTotal)
			for i, r := range totalResults {
				if r.UserName == newT.UserName {
					totalResults[i] = newT
					return
				}
			}
			totalResults = append(totalResults, newT)
		},
		UpdateFunc: func(oldObj, newObj any) {
			newT := newObj.(UserTotal)
			for i, r := range totalResults {
				if r.UserName == newT.UserName {
					totalResults[i] = newT
					return
				}
			}
			totalResults = append(totalResults, newT)
		},
	})

	// Pushing data as pointers to ensure unique identity
	users.OnAdd(&User{ID: 1, Name: "Alice"}, true)
	orders.OnAdd(&Order{ID: 101, UserID: 1, Amount: 50}, true)
	orders.OnAdd(&Order{ID: 102, UserID: 1, Amount: 30}, false)

	if len(totalResults) != 1 {
		t.Errorf("Expected 1 total result, got %d", len(totalResults))
	} else if totalResults[0].Total != 80 {
		t.Errorf("Expected total 80, got %d", totalResults[0].Total)
	}
}

func TestFlatMap(t *testing.T) {
	type TaggedItem struct {
		ID  int
		Tag string
	}
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	m := QueryInformer(&FlatMap[*TaggedItem, int]{
		Lock: lock,
		Map: func(i int) ([]*TaggedItem, error) {
			return []*TaggedItem{
				{ID: i, Tag: "even"},
				{ID: i, Tag: "odd"},
			}, nil
		},
		Over: source,
	})

	var results []*TaggedItem
	m.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			results = append(results, obj.(*TaggedItem))
		},
	})

	source.OnAdd(1, true)
	if len(results) != 2 {
		t.Errorf("Expected 2 tagged items, got %d", len(results))
	}
}

func TestJoinUpdatesAndDeletes(t *testing.T) {
	lock := NewLockGroup()
	left := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)
	right := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	joined := QueryInformer(&Join[UserOrder, *User, *Order]{
		Lock: lock,
		Select: func(u *User, o *Order) (UserOrder, error) {
			return UserOrder{UserName: u.Name, Amount: o.Amount}, nil
		},
		From: left,
		Join: right,
		On: func(u *User, o *Order) any {
			if u != nil {
				return [1]int{u.ID}
			}
			return [1]int{o.UserID}
		},
		Where: func(u *User, o *Order) bool {
			return u.ID == o.UserID
		},
	})

	var results []UserOrder
	joined.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) { results = append(results, obj.(UserOrder)) },
		UpdateFunc: func(old, new any) {
			for i, r := range results {
				if r.UserName == old.(UserOrder).UserName && r.Amount == old.(UserOrder).Amount {
					results[i] = new.(UserOrder)
					return
				}
			}
		},
		DeleteFunc: func(obj any) {
			val := obj.(UserOrder)
			for i, r := range results {
				if r.UserName == val.UserName && r.Amount == val.Amount {
					results = append(results[:i], results[i+1:]...)
					return
				}
			}
		},
	})

	u1 := &User{ID: 1, Name: "Alice"}
	o1 := &Order{ID: 101, UserID: 1, Amount: 50}
	o2 := &Order{ID: 102, UserID: 1, Amount: 100}

	left.OnAdd(u1, true)
	right.OnAdd(o1, true)
	right.OnAdd(o2, false)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Update order amount
	right.OnUpdate(o1, &Order{ID: 101, UserID: 1, Amount: 60})
	found := false
	for _, r := range results {
		if r.Amount == 60 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Updated amount 60 not found in results")
	}

	// Delete user -> should delete all joined results
	left.OnDelete(u1)
	if len(results) != 0 {
		t.Errorf("Expected 0 results after left delete, got %d", len(results))
	}
}

func TestGroupByAggregations(t *testing.T) {
	lock := NewLockGroup()
	source := NewManualSharedInformerWithOptions(lock, cache.MetaNamespaceKeyFunc)

	type CategoryTotal struct {
		Category string
		Sum      int64
		Count    int64
	}

	query := QueryInformer(&GroupBy[CategoryTotal, *UserOrder]{
		Lock: lock,
		Select: func(fields []GroupField) (CategoryTotal, error) {
			return CategoryTotal{
				Category: fields[0].(string),
				Sum:      fields[1].(int64),
				Count:    fields[2].(int64),
			}, nil
		},
		From: source,
		GroupBy: func(uo *UserOrder) (any, []GroupField) {
			return [1]string{uo.UserName},
				[]GroupField{
					AnyValue(uo.UserName),
					Sum(int64(uo.Amount)),
					Count(),
				}
		},
	})

	var latest CategoryTotal
	query.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { latest = obj.(CategoryTotal) },
		UpdateFunc: func(old, new any) { latest = new.(CategoryTotal) },
	})

	source.OnAdd(&UserOrder{UserName: "Alice", Amount: 10}, true)
	source.OnAdd(&UserOrder{UserName: "Alice", Amount: 20}, false)

	if latest.Sum != 30 || latest.Count != 2 {
		t.Errorf("Expected Sum 30 Count 2, got %+v", latest)
	}
}
