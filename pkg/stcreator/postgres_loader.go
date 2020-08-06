package stcreator

import (
	"database/sql"

	"github.com/achiku/varfmt"
	"github.com/iancoleman/strcase"
	_ "github.com/lib/pq"
)

const tablesSQL = `
SELECT c.relkind AS type, c.relname AS table_name
FROM pg_class c
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1
AND c.relkind = 'r'
ORDER BY c.relname;
`

const columnsTableSQL = `
SELECT
    a.attname AS name,
    a.attnotnull AS not_null,    
    COALESCE(ct.contype = 'p', false) AS  is_primary_key,
    COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') AS default_value,
    CASE
        WHEN a.atttypid = ANY ('{int,int8,int2}'::regtype[])
          AND EXISTS (
             SELECT 1 FROM pg_attrdef ad
             WHERE  ad.adrelid = a.attrelid
             AND    ad.adnum   = a.attnum
             AND    ad.adbin = 'nextval('''
                || (pg_get_serial_sequence (a.attrelid::regclass::text
                                          , a.attname))::regclass
                || '''::regclass)'
             )
            THEN CASE a.atttypid
                    WHEN 'int'::regtype  THEN 'serial'
                    WHEN 'int8'::regtype THEN 'bigserial'
                    WHEN 'int2'::regtype THEN 'smallserial'
                 END
        WHEN a.atttypid = ANY ('{uuid}'::regtype[]) AND COALESCE(pg_get_expr(ad.adbin, ad.adrelid), '') != ''
            THEN 'autogenuuid'
        ELSE format_type(a.atttypid, a.atttypmod)
    END AS data_type
FROM pg_attribute a
JOIN ONLY pg_class c ON c.oid = a.attrelid
JOIN ONLY pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_constraint ct ON ct.conrelid = c.oid
AND a.attnum = ANY(ct.conkey) AND ct.contype = 'p'
LEFT JOIN pg_attrdef ad ON ad.adrelid = c.oid AND ad.adnum = a.attnum
WHERE a.attisdropped = false
AND n.nspname = $1
AND c.relname = $2
AND a.attnum > 0
ORDER BY a.attnum;
`

type MapType struct {
	Type, NullType string
	DBTypes        []string
}

type MapTypes []MapType

func (m MapTypes) At(t string) (MapType, bool) {
	for _, mapType := range m {
		for _, dbType := range mapType.DBTypes {
			if t == dbType {
				return mapType, true
			}
		}
	}
	return MapType{}, false
}

var mapTypesPkg = map[string]string{
	"uuid.UUID":       "github.com/google/uuid",
	"*uuid.UUID":      "github.com/google/uuid",
	"sql.NullString":  "database/sql",
	"sql.NullInt64":   "database/sql",
	"sql.NullFloat64": "database/sql",
	"time.Duration":   "time",
	"time.Time":       "time",
	"*time.Duration":  "time",
	"*time.Time":      "time",
}

var mapTypes = MapTypes{
	{
		Type:     "uuid.UUID",
		NullType: "*uuid.UUID",
		DBTypes:  []string{"uuid"},
	},
	{
		Type:     "string",
		NullType: "sql.NullString",
		DBTypes:  []string{"character", "character varying", "text", "money"},
	},
	{
		Type:     "time.Time",
		NullType: "*time.Time",
		DBTypes:  []string{"time with time zone", "time without time zone", "timestamp without time zone", "timestamp with time zone", "date"},
	},
	{
		Type:     "bool",
		NullType: "bool",
		DBTypes:  []string{"boolean"},
	},
	{
		Type:     "int16",
		NullType: "sql.NullInt64",
		DBTypes:  []string{"smallint"},
	},
	{
		Type:     "int",
		NullType: "sql.NullInt64",
		DBTypes:  []string{"integer"},
	},
	{
		Type:     "int64",
		NullType: "sql.NullInt64",
		DBTypes:  []string{"bigint"},
	},
	{
		Type:     "uint16",
		NullType: "sql.NullInt64",
		DBTypes:  []string{"smallserial"},
	},
	{
		Type:     "uint32",
		NullType: "sql.NullInt64",
		DBTypes:  []string{"serial"},
	}, {
		Type:     "float32",
		NullType: "sql.NullFloat64",
		DBTypes:  []string{"real"},
	},
	{
		Type:     "float64",
		NullType: "sql.NullFloat64",
		DBTypes:  []string{"numeric", "double precision"},
	},
	{
		Type:     "byte",
		NullType: "byte",
		DBTypes:  []string{"bytea"},
	},
	{
		Type:     "[]byte",
		NullType: "[]byte",
		DBTypes:  []string{"json", "jsonb"},
	},
	{
		Type:     "[]byte",
		NullType: "[]byte",
		DBTypes:  []string{"xml"},
	},
	{
		Type:     "time.Duration",
		NullType: "*time.Duration",
		DBTypes:  []string{"interval"},
	},
	{
		Type:     "[]int",
		NullType: "[]int",
		DBTypes:  []string{"integer[]"},
	},
	{
		Type:     "[]string",
		NullType: "[]string",
		DBTypes:  []string{"string[]"},
	},
}

type pgTable struct {
	Name string
	Type string
}

type pgTableParam struct {
	Name    string
	Type    string
	NotNull bool
	Primary bool
	Default string
}

type PostgresLoader struct {
	URL    string   `yaml:"url"`
	Tables []string `yaml:"tables"`
}

func (*PostgresLoader) Name() string {
	return "postgres"
}

func (l *PostgresLoader) Process() (result []StructMetadata, err error) {
	conn, err := sql.Open("postgres", l.URL)
	if err != nil {
		return result, err
	}
	rows, err := conn.Query(tablesSQL, "public")
	if err != nil {
		return result, err
	}
	tables := map[string]*pgTable{}
	for rows.Next() {
		t := &pgTable{}
		err := rows.Scan(&t.Type, &t.Name)
		if err != nil {
			return result, err
		}
		tables[t.Name] = t
	}
	for _, table := range l.Tables {
		if t, ok := tables[table]; ok {
			name := t.Name
			if name[len(name)-1] == 's' {
				name = name[:len(name)-1]
			}
			sm := StructMetadata{
				Name:      varfmt.PublicVarName(name),
				LowerName: strcase.ToLowerCamel(name),
			}
			rows, err := conn.Query(columnsTableSQL, "public", t.Name)
			if err != nil {
				return result, err
			}

			existsPkgs := map[string]struct{}{}

			for rows.Next() {
				p := &pgTableParam{}
				err := rows.Scan(&p.Name, &p.NotNull, &p.Primary, &p.Default, &p.Type)
				if err != nil {
					return result, err
				}
				mt, ok := mapTypes.At(p.Type)
				if !ok {
					mt = MapType{
						Type:     "interface{}",
						NullType: "interface{}",
					}
				}
				sp := StructParam{
					Name:       varfmt.PublicVarName(p.Name),
					LowerName:  strcase.ToLowerCamel(p.Name),
					RawType:    mt.Type,
					ColumnName: p.Name,
					Primary:    p.Primary,
					NotNull:    p.NotNull,
					Default:    p.Default,
				}
				if sp.NotNull {
					sp.Type = mt.Type
				} else {
					sp.Type = mt.NullType
				}
				if sp.Primary {
					sm.Primary = sp
				}
				if pkg, ok := mapTypesPkg[sp.Type]; ok {
					if _, ok := existsPkgs[pkg]; !ok {
						sm.Imports = append(sm.Imports, StructImport{
							Pkg:   pkg,
							Param: sp,
						})
						existsPkgs[pkg] = struct{}{}
					}
				}
				sm.Params = append(sm.Params, sp)
			}
			result = append(result, sm)
		}
	}
	return
}
