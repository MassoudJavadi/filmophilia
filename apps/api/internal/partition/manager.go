package partition

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds partition manager configuration
type Config struct {
	// MonthsAhead specifies how many months ahead to create partitions
	MonthsAhead int
	// CheckInterval specifies how often to check for missing partitions
	CheckInterval time.Duration
}

// DefaultConfig returns the default partition manager configuration
func DefaultConfig() Config {
	return Config{
		MonthsAhead:   3,              // Create partitions 3 months ahead
		CheckInterval: 24 * time.Hour, // Check daily
	}
}

// Manager handles automatic partition creation for high-growth tables
type Manager struct {
	db     *pgxpool.Pool
	config Config
	stop   chan struct{}
	done   chan struct{}
}

// NewManager creates a new partition manager
func NewManager(db *pgxpool.Pool, config Config) *Manager {
	return &Manager{
		db:     db,
		config: config,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Start begins the background partition management
func (m *Manager) Start(ctx context.Context) {
	// Run initial check on startup
	if err := m.EnsurePartitions(ctx); err != nil {
		log.Printf("Partition manager: initial check failed: %v", err)
	}

	go m.runLoop(ctx)
}

// Stop gracefully stops the partition manager
func (m *Manager) Stop() {
	close(m.stop)
	<-m.done
}

func (m *Manager) runLoop(ctx context.Context) {
	defer close(m.done)

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stop:
			log.Println("Partition manager: stopped")
			return
		case <-ctx.Done():
			log.Println("Partition manager: context cancelled")
			return
		case <-ticker.C:
			if err := m.EnsurePartitions(ctx); err != nil {
				log.Printf("Partition manager: check failed: %v", err)
			}
		}
	}
}

// EnsurePartitions checks and creates any missing partitions for upcoming months
func (m *Manager) EnsurePartitions(ctx context.Context) error {
	now := time.Now()

	for i := 0; i <= m.config.MonthsAhead; i++ {
		targetMonth := now.AddDate(0, i, 0)
		year := targetMonth.Year()
		month := int(targetMonth.Month())

		for _, table := range []string{"notifications", "activities"} {
			exists, err := m.partitionExists(ctx, table, year, month)
			if err != nil {
				return fmt.Errorf("checking partition %s_%d_%02d: %w", table, year, month, err)
			}

			if !exists {
				if err := m.createPartition(ctx, table, year, month); err != nil {
					return fmt.Errorf("creating partition %s_%d_%02d: %w", table, year, month, err)
				}
				log.Printf("Partition manager: created %s_%d_%02d", table, year, month)
			}
		}
	}

	return nil
}

// partitionExists checks if a partition table exists
func (m *Manager) partitionExists(ctx context.Context, table string, year, month int) (bool, error) {
	partitionName := fmt.Sprintf("%s_%d_%02d", table, year, month)

	query := `
		SELECT EXISTS (
			SELECT 1 FROM pg_tables
			WHERE schemaname = 'public'
			AND tablename = $1
		)
	`

	var exists bool
	err := m.db.QueryRow(ctx, query, partitionName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// createPartition creates a new monthly partition for the specified table
func (m *Manager) createPartition(ctx context.Context, table string, year, month int) error {
	partitionName := fmt.Sprintf("%s_%d_%02d", table, year, month)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
		FOR VALUES FROM ('%s') TO ('%s')
	`, partitionName, table, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	_, err := m.db.Exec(ctx, query)
	return err
}

// ListPartitions returns all existing partitions for a table
func (m *Manager) ListPartitions(ctx context.Context, table string) ([]string, error) {
	query := `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename LIKE $1
		ORDER BY tablename
	`

	rows, err := m.db.Query(ctx, query, table+"_%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partitions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		partitions = append(partitions, name)
	}

	return partitions, rows.Err()
}

// GetPartitionStats returns statistics about partition coverage
func (m *Manager) GetPartitionStats(ctx context.Context) (map[string]PartitionStats, error) {
	stats := make(map[string]PartitionStats)

	for _, table := range []string{"notifications", "activities"} {
		partitions, err := m.ListPartitions(ctx, table)
		if err != nil {
			return nil, err
		}

		tableStats := PartitionStats{
			TableName:      table,
			PartitionCount: len(partitions),
			Partitions:     partitions,
		}

		// Find earliest and latest partitions (excluding default)
		for _, p := range partitions {
			if p == table+"_default" {
				continue
			}
			if tableStats.Earliest == "" || p < tableStats.Earliest {
				tableStats.Earliest = p
			}
			if tableStats.Latest == "" || p > tableStats.Latest {
				tableStats.Latest = p
			}
		}

		stats[table] = tableStats
	}

	return stats, nil
}

// PartitionStats holds statistics about a partitioned table
type PartitionStats struct {
	TableName      string
	PartitionCount int
	Partitions     []string
	Earliest       string
	Latest         string
}
