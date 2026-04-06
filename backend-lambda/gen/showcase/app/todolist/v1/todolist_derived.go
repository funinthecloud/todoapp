package todolistv1

// AfterOn computes derived fields from the items collection.
// Called after event replay (Load) and after materialization (Apply).
func (t *TodoList) AfterOn() {
	t.ItemCount = int32(len(t.Items))
	var completed int32
	for _, item := range t.Items {
		if item.GetCompleted() {
			completed++
		}
	}
	t.CompletedCount = completed
}
