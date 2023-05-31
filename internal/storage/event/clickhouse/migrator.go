package clickhouse

//type Migrator struct {
//	migrator *chmigrate.Migrator
//}
//
//func NewMigrator(cfg Config) (*Migrator, error) {
//	db := ch.Connect(
//		ch.WithDSN(fmt.Sprintf("clickhouse://%s/%s?sslmode=disable", cfg.Addr, cfg.DB)),
//		ch.WithUser(cfg.Username),
//		ch.WithPassword(cfg.Password),
//	)
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
//	defer cancel()
//	if err := db.Ping(ctx); err != nil {
//		return nil, err
//	}
//
//	return &Migrator{
//		migrator: chmigrate.NewMigrator(db, migrations.Migrations),
//	}, nil
//}
//
//func (m *Migrator) MigrateDB(ctx context.Context) error {
//	if err := m.migrator.Lock(ctx); err != nil {
//		return err
//	}
//	defer m.migrator.Unlock(ctx)
//
//	group, err := m.migrator.Migrate(ctx)
//	if err != nil {
//		return err
//	}
//
//	if group.IsZero() {
//		log.Info().Msg("there are no new migrations to run (database is up to date)")
//		return nil
//	}
//	log.Info().Msg(fmt.Sprintf("migrated to %s", group))
//
//	return nil
//}
