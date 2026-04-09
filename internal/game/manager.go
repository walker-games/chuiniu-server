package game

import (
	"sync"
	"time"
)

type RoomManager struct {
	rooms       map[string]*Room  // room ID → Room
	codeToRoom  map[string]string // invite code → room ID
	mu          sync.RWMutex
	idleTimeout time.Duration
}

func NewRoomManager(idleTimeout time.Duration) *RoomManager {
	m := &RoomManager{
		rooms:       make(map[string]*Room),
		codeToRoom:  make(map[string]string),
		idleTimeout: idleTimeout,
	}
	go m.cleanupLoop()
	return m
}

func (m *RoomManager) CreateRoom(hostID, hostName, hostAvatar string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	room := NewRoom(hostID, hostName, hostAvatar)

	// Ensure unique room ID
	for {
		if _, exists := m.rooms[room.ID]; !exists {
			break
		}
		room.ID = generateRoomID()
	}

	// Ensure unique invite code
	for {
		if _, exists := m.codeToRoom[room.Code]; !exists {
			break
		}
		room.Code = generateInviteCode()
	}

	m.rooms[room.ID] = room
	m.codeToRoom[room.Code] = room.ID
	return room
}

func (m *RoomManager) GetRoom(roomID string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.rooms[roomID]
}

func (m *RoomManager) GetRoomByCode(code string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	roomID, ok := m.codeToRoom[code]
	if !ok {
		return nil
	}
	return m.rooms[roomID]
}

func (m *RoomManager) RemoveRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, ok := m.rooms[roomID]; ok {
		delete(m.codeToRoom, room.Code)
		delete(m.rooms, roomID)
	}
}

func (m *RoomManager) RoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.rooms)
}

func (m *RoomManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, room := range m.rooms {
			room.mu.RLock()
			idle := now.Sub(room.LastActive) > m.idleTimeout
			empty := room.ConnectedCountLocked() == 0
			room.mu.RUnlock()

			if idle || empty {
				delete(m.codeToRoom, room.Code)
				delete(m.rooms, id)
			}
		}
		m.mu.Unlock()
	}
}
