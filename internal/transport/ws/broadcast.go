package ws

type BroadcastMessage struct {
	userIDs []int64
	message []byte
}