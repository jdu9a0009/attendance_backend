package commands

import (
	"attendance/backend/internal/pkg/repository/postgresql"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

type Scheme struct {
	Index       int
	Description string
	Query       string
}

var scheme = []Scheme{

	{
		Index:       1,
		Description: "CREATE TYPE \"user_role\" AS ENUM",
		Query: `
        CREATE TYPE "user_role" AS ENUM ('EMPLOYEE', 'ADMIN','DASHBOARD','QRCODE');`,
	},
	{
		Index:       2,
		Description: "Create table: users.",
		Query: `
        CREATE TABLE IF NOT EXISTS users (
            id serial primary key,
            employee_id text not null,
            password text not null,
            role user_role,
            first_name text,
			last_name text,
            created_at timestamp default now(),
            created_by int references users(id),
            updated_at timestamp,
            updated_by int references users(id),
            deleted_at timestamp,
            deleted_by int references users(id)
        );`,
	},
	{
		Index:       3,
		Description: "Create area with employee_id: Admin01, password: 1",
		Query: `
        INSERT INTO users(employee_id, role, password)
        SELECT 'Admin01', 'ADMIN', '$2a$10$NKtnMwDPFSQLG6uOi4Zqheru5Ygbj9TWFHjpl478rRSaO5cJ9QuH2'
        WHERE NOT EXISTS (SELECT employee_id FROM users WHERE employee_id = 'Admin01');
        `,
	},
	{
		Index:       4,
		Description: "Create area with employee_id: QrCode01, password: 1",
		Query: `
        INSERT INTO users(employee_id, role, password)
        SELECT 'QrCode01', 'QRCODE', '$2a$10$NKtnMwDPFSQLG6uOi4Zqheru5Ygbj9TWFHjpl478rRSaO5cJ9QuH2'
        WHERE NOT EXISTS (SELECT employee_id FROM users WHERE employee_id = 'QrCode01');
        `,
	},
	{
		Index:       5,
		Description: "Create area with employee_id: Dashboard01, password: 1",
		Query: `
        INSERT INTO users(employee_id, role, password)
        SELECT 'Dashboard01', 'DASHBOARD', '$2a$10$NKtnMwDPFSQLG6uOi4Zqheru5Ygbj9TWFHjpl478rRSaO5cJ9QuH2'
        WHERE NOT EXISTS (SELECT employee_id FROM users WHERE employee_id = 'Dashboard01');
        `,
	},
	{
		Index:       6,
		Description: "Create table: department",
		Query: `
        CREATE TABLE IF NOT EXISTS department (
            id serial primary key,
            name text not null,
			display_number int not null,
			department_nickname text ,
            created_at timestamp default now(),
            created_by int references users(id),
            updated_at timestamp,
            updated_by int references users(id),
            deleted_at timestamp,
            deleted_by int references users(id)
        );`,
	},
	{
		Index:       7,
		Description: "Create table: position.",
		Query: `
        CREATE TABLE IF NOT EXISTS position (
            id serial primary key,
            name text not null,
            department_id int references department(id),
            created_at timestamp default now(),
            created_by int references users(id),
            updated_at timestamp,
            updated_by int references users(id),
            deleted_at timestamp,
            deleted_by int references users(id)
        );`,
	},
	{
		Index:       8,
		Description: "Alter table users",
		Query: `
        ALTER TABLE users
		ADD COLUMN IF NOT EXISTS nick_name varchar(255),
        ADD COLUMN IF NOT EXISTS department_id int references department(id),
		ADD COLUMN IF NOT EXISTS position_id int references position(id),
		 ADD COLUMN IF NOT EXISTS phone VARCHAR(255),
        ADD COLUMN IF NOT EXISTS status BOOLEAN DEFAULT false,
        ADD COLUMN IF NOT EXISTS email VARCHAR(255);`,
	},

	{
		Index:       9,
		Description: "Create table: attendance.",
		Query: `
        CREATE TABLE attendance (
            id SERIAL PRIMARY KEY,
            employee_id VARCHAR NOT NULL,
            come_time TIME NOT NULL,
            work_day DATE NOT NULL,
            leave_time TIME,
            status BOOLEAN DEFAULT true,
		    forget_leave BOOLEAN DEFAULT false,
            created_at TIMESTAMP DEFAULT NOW(),
            created_by INT REFERENCES users(id),
            updated_at TIMESTAMP,
            updated_by INT REFERENCES users(id),
            deleted_at TIMESTAMP,
            deleted_by INT REFERENCES users(id)
        );`,
	},
	{
		Index:       10,
		Description: "Create table: attendance_period.",
		Query: `
        CREATE TABLE attendance_period (
            id SERIAL PRIMARY KEY,
            attendance_id  int NOT NULL REFERENCES attendance(id),
            come_time TIME NOT NULL,
			leave_time TIME,
			updated_at TIMESTAMP,
            work_day DATE NOT NULL
        );`,
	},
	{
		Index:       11,
		Description: "Create table: company_info.",
		Query: `
        CREATE TABLE company_info (
            id SERIAL PRIMARY KEY,
            company_name VARCHAR(250) NOT NULL,
			url VARCHAR(100),
			latitude FLOAT NOT NULL,
			longitude FLOAT NOT NULL,
			radius FLOAT NOT NULL,
            start_time TIME,
			end_time TIME,
			late_time TIME,
			over_end_time TIME,
			come_time_color varchar(200),
			leave_time_color varchar(200),
			forget_time_color varchar(200),
			present_color varchar(200),
			absent_color varchar(200),
			new_present_color varchar(200),
			new_absent_color varchar(200),
            created_at TIMESTAMP DEFAULT NOW(),
            created_by INT REFERENCES users(id),
            updated_at TIMESTAMP,
            updated_by INT REFERENCES users(id),
			deleted_at TIMESTAMP,
            deleted_by INT REFERENCES users(id)
        );`,
	},
	{
		Index:       12,
		Description: "Insert data fortable: company_info.",
		Query: `
        INSERT INTO company_info (
        id,
        company_name,
        url,
        latitude,
        longitude,
		radius,
        start_time,
        end_time,
        late_time,
        over_end_time,
		come_time_color,
		leave_time_color,
		forget_time_color,
		present_color,
		absent_color,
		new_present_color,
		new_absent_color,
        created_by,
        updated_by
    ) VALUES (
        1,
        'Digital Knowledge',
        'statics/company_info/2024-09-24T20:49:17+05:00-Screenshot from 2024-09-24 13-55-14.png',
        35.7031509,
        139.7745439,
		3000.0,
        '09:00:00',
        '18:00:00',
        '09:20:00',
        '22:30:00',
		'#e33935',
		'#fbfbfc',
		'#f8f79e',
		'#7b6bff',
		'#1e67f4',
		'#34bba8',
		'#ebcb58',
        1,
        1
);`,
	},
	{
		Index:       13,
		Description: "Creating Trigger and Function fo Websocket",
		Query: `CREATE OR REPLACE FUNCTION notify_attendance_change()
                   RETURNS TRIGGER AS $$
                   BEGIN
                     PERFORM pg_notify('attendance_changes', json_build_object(
                             'operation', TG_OP,
                             'data', row_to_json(NEW)
                         )::text);
                         RETURN NEW;
                     END;
                     $$ LANGUAGE plpgsql;

                CREATE TRIGGER attendance_changes_trigger
                AFTER INSERT OR UPDATE ON attendance_period
                FOR EACH ROW EXECUTE FUNCTION notify_attendance_change();`,
	},
}

// Migrate creates the scheme in the database.
func Migrate(db *postgresql.Database) {
	for _, s := range scheme {
		if _, err := db.Query(s.Query); err != nil {
			log.Fatalln("migrate error", err)
		}
	}
}

func MigrateUP(db *postgresql.Database) {
	var (
		version int
		dirty   bool
		er      *string
	)
	err := db.QueryRow("SELECT version, dirty, error FROM schema_migrations").Scan(&version, &dirty, &er)
	if err != nil {
		if err.Error() == `ERROR: relation "schema_migrations" does not exist (SQLSTATE=42P01)` {
			if _, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS schema_migrations (version int not null, dirty bool not null, error text);
				DELETE FROM schema_migrations;
				INSERT INTO schema_migrations (version, dirty) values (0, false);
			`); err != nil {
				log.Fatalln("migrate schema_migrations create error", err)
			}
			version = 0
			dirty = false
		} else {
			log.Fatalln("migrate schema_migrations scan: ", err)
		}
	}

	if dirty {
		for _, v := range scheme {
			if v.Index == version {
				if _, err = db.Exec(v.Query); err != nil {
					if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET error = '%s'`, err.Error())); err != nil {
						log.Fatalln("migrate error", err)
					}
					log.Fatalln(fmt.Sprintf("migrate error version: %d", version), err)
				}
				if _, err = db.Exec(`UPDATE schema_migrations SET dirty = false, error = null`); err != nil {
					log.Fatalln("migrate error", err)
				}
			}
		}
	}

	for _, s := range scheme {
		if s.Index > version {
			if _, err = db.Exec(s.Query); err != nil {
				if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET error = '%s', version = %d, dirty = true`, err.Error(), s.Index)); err != nil {
					log.Fatalln("migrate error", err)
				}
				log.Fatalln(fmt.Sprintf("migrate error version: %d", s.Index), err)
			}
			if _, err = db.Exec(fmt.Sprintf(`UPDATE schema_migrations SET version = %d`, s.Index)); err != nil {
				log.Fatalln("migrate error", err)
			}
		}
	}
}
