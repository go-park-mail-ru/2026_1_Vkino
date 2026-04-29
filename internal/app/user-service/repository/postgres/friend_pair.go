package postgres

// orderedFriendPair возвращает (user1_id, user2_id) для таблицы friend, где user1_id <= user2_id.
func orderedFriendPair(a, b int64) (int64, int64) {
	if a > b {
		return b, a
	}

	return a, b
}
