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
	GuildD    string    `db:"guildid"`
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

func convertToDomain(modelVoice dbVoiceData) domain.VoiceData {
	return domain.VoiceData{
		GuildID:   modelVoice.GuildD,
		ID:        modelVoice.ID,
		Timestamp: modelVoice.Timestamp,
		Name:      modelVoice.Name,
		UserID:    modelVoice.UserID,
		Duration:  modelVoice.Duration,
	}
}

func (v VoiceDataRepository) Save(data domain.VoiceData) error {
	model := convertToModel(data)
	_, err := v.db.NamedExec("INSERT INTO voicedata (id, guildid, timestamp,name,userid,duration) "+
		"VALUES (:id, :guildid, :timestamp, :name, :userid, :duration)", model)
	if err != nil {
		return fmt.Errorf("err save voice data: %w", err)
	}
	return nil
}

func (v VoiceDataRepository) GetOnRange(guildID string, from time.Time, to time.Time) ([]domain.VoiceData, error) {
	var rows []dbVoiceData
	query := "select * from voicedata v where guildid = $1 and timestamp >= $2 and timestamp <= $3"
	if err := v.db.Select(&rows, query, guildID, from, to); err != nil {
		return nil, fmt.Errorf("err get voice data on range:%w", err)
	}
	models := make([]domain.VoiceData, len(rows))
	for i, row := range rows {
		models[i] = convertToDomain(row)
	}
	return models, nil
}

func NewVoiceDataRepository(db *sqlx.DB) *VoiceDataRepository {
	return &VoiceDataRepository{db: db}
}
