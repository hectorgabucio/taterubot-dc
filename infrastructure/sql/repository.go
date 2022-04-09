package sqlrepo

import (
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/jmoiron/sqlx"
	"time"
)

type VoiceDataRepository struct {
	db *sqlx.DB
}

type dbVoiceData struct {
	ID        string    `db:"id"`
	GuildD    string    `db:"guildId"`
	Timestamp time.Time `db:"timestamp"`
	Name      string    `db:"name"`
	UserID    string    `db:"userid"`
	Duration  int       `db:"duration"`
}

func convertToModel(domainVoice domain.VoiceData) dbVoiceData {
	return dbVoiceData{
		ID:        domainVoice.ID,
		GuildD:    domainVoice.GuildID,
		Timestamp: domainVoice.Timestamp,
		Name:      domainVoice.Name,
		UserID:    domainVoice.UserID,
		Duration:  domainVoice.Duration,
	}
}

func (v VoiceDataRepository) Save(data domain.VoiceData) error {
	model := convertToModel(data)
	_, err := v.db.NamedExec("INSERT INTO voicedata (id, guildid, timestamp,name,userid,duration) "+
		"VALUES (:id, :guildId, :timestamp, :name, :userid, :duration)", model)
	if err != nil {
		return fmt.Errorf("err save voice data: %w", err)
	}
	return nil
}

func (v VoiceDataRepository) GetOnRange(guildID string, from time.Time, to time.Time) []domain.VoiceData {
	//TODO implement me
	panic("implement me")
}

func NewVoiceDataRepository(db *sqlx.DB) *VoiceDataRepository {
	return &VoiceDataRepository{db: db}
}
