// Local-only, insert-only FoodHub presentation seed. Default execution is offline.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"alpha-dev/models"
	"alpha-dev/utils"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

//go:embed catalog.json
var catalogJSON []byte

const databaseName = "foodhub_new_structure"
const seedMarker = "foodhub-local-presentation-seed:v1"

type dish struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Available   bool   `json:"available"`
}
type restaurant struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	Phone       string `json:"phone"`
	OpeningTime string `json:"opening_time"`
	ClosingTime string `json:"closing_time"`
	ImageURL    string `json:"image_url"`
	Items       []dish `json:"items"`
}
type record struct {
	Kind     string `gorm:"column:kind"`
	SeedKey  string `gorm:"column:seed_key"`
	EntityID uint   `gorm:"column:entity_id"`
}
type connection struct{ Host, Port, Name, User, Password, SSLMode string }
type stats struct{ Managers, Categories, Restaurants, Items, Reused int }

func loadCatalog() ([]restaurant, error) {
	var rows []restaurant
	decoder := json.NewDecoder(bytes.NewReader(catalogJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rows); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("catalog has trailing JSON")
	}
	return rows, validateCatalog(rows)
}
func validateCatalog(rows []restaurant) error {
	if len(rows) != 6 {
		return errors.New("catalog must contain six restaurants")
	}
	names, keys := map[string]bool{}, map[string]bool{}
	for _, r := range rows {
		if r.Key == "" || strings.ContainsAny(r.Key, " /@") || keys[r.Key] || r.Name == "" || names[r.Name] {
			return errors.New("duplicate/invalid restaurant identity")
		}
		keys[r.Key], names[r.Name] = true, true
		if !validImagePath(r.ImageURL) || r.Description == "" || r.Address == "" || r.City == "" || r.State == "" || r.Phone == "" {
			return fmt.Errorf("incomplete restaurant: %s", r.Key)
		}
		opening, e1 := time.Parse("15:04", r.OpeningTime)
		closing, e2 := time.Parse("15:04", r.ClosingTime)
		if e1 != nil || e2 != nil || !closing.After(opening) {
			return fmt.Errorf("invalid hours: %s", r.Key)
		}
		if len(r.Items) < 6 || len(r.Items) > 8 {
			return fmt.Errorf("expected 6–8 menu items: %s", r.Key)
		}
		itemKeys, itemNames, available := map[string]bool{}, map[string]bool{}, 0
		for _, d := range r.Items {
			if d.Key == "" || itemKeys[d.Key] || d.Name == "" || itemNames[d.Name] || (d.ImageURL != "" && !validImagePath(d.ImageURL)) || d.Description == "" || d.Category == "" || d.Price <= 0 || d.Price > 10000 {
				return fmt.Errorf("invalid/duplicate dish in %s", r.Key)
			}
			itemKeys[d.Key], itemNames[d.Name] = true, true
			if d.Available {
				available++
			}
		}
		if available < 6 {
			return fmt.Errorf("at least six orderable dishes required: %s", r.Key)
		}
	}
	return nil
}
func validImagePath(value string) bool {
	return strings.HasPrefix(value, "/demo-images/") && (strings.HasSuffix(value, ".png") || (strings.HasPrefix(value, "/demo-images/photos/") && strings.HasSuffix(value, ".jpg"))) && !strings.Contains(value, "..") && !strings.ContainsAny(value, "?\\#")
}
func validateImageFiles(rows []restaurant) error {
	for _, r := range rows {
		paths := []string{r.ImageURL}
		for _, d := range r.Items {
			paths = append(paths, d.ImageURL)
		}
		for _, value := range paths {
			if value == "" {
				continue
			}
			info, err := os.Stat(filepath.Join("frontend/public", strings.TrimPrefix(value, "/")))
			if err != nil || info.IsDir() || info.Size() == 0 {
				return fmt.Errorf("missing local image: %s; generate images before applying seed", value)
			}
		}
	}
	return nil
}
func validateConnection(c connection) error {
	if c.Name != databaseName {
		return errors.New("refusing a database other than LOCAL foodhub_new_structure")
	}
	if c.Host != "localhost" && c.Host != "127.0.0.1" && c.Host != "::1" {
		return errors.New("refusing a non-loopback database host")
	}
	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 || c.User == "" {
		return errors.New("invalid database port/user")
	}
	switch c.SSLMode {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return errors.New("invalid SSL_MODE")
	}
	return nil
}
func readConnection() (connection, error) {
	// Read literal project settings without changing process-wide environment or
	// accepting PGDATABASE/PGHOST/PGSERVICE overrides from a developer's shell.
	values, err := godotenv.Read(".env")
	if err != nil {
		return connection{}, errors.New("cannot read project-root .env")
	}
	c := connection{values["DB_HOST"], values["DB_PORT"], values["DB_NAME"], values["DB_USER"], values["DB_PASSWORD"], values["SSL_MODE"]}
	for _, name := range []string{"APP_ENV", "GO_ENV", "ENVIRONMENT"} {
		for _, value := range []string{values[name], os.Getenv(name)} {
			if value != "" && value != "development" && value != "dev" && value != "local" {
				return connection{}, errors.New("refusing a non-development environment")
			}
		}
	}
	return c, validateConnection(c)
}
func (c connection) DSN() string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(c.Host, c.Port), Path: "/" + c.Name, User: url.UserPassword(c.User, c.Password)}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	q.Set("connect_timeout", "5")
	q.Set("options", "-c search_path=public -c lock_timeout=5000 -c statement_timeout=30000")
	u.RawQuery = q.Encode()
	return u.String()
}
func main() {
	apply := flag.Bool("apply", false, "write the reviewed seed to the LOCAL development database")
	confirm := flag.String("confirm-local", "", "must equal foodhub_new_structure when --apply is used")
	flag.Parse()
	rows, err := loadCatalog()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !*apply {
		fmt.Println("OFFLINE PREVIEW ONLY: no database connection or writes.")
		for _, r := range rows {
			available := 0
			for _, d := range r.Items {
				if d.Available {
					available++
				}
			}
			fmt.Printf("%s: %d dishes (%d available)\n", r.Name, len(r.Items), available)
		}
		fmt.Println("After approval: go run ./tools/seed-foodhub --apply --confirm-local foodhub_new_structure")
		return
	}
	if *confirm != databaseName {
		fmt.Fprintln(os.Stderr, "Missing exact --confirm-local foodhub_new_structure; no connection made.")
		os.Exit(1)
	}
	if err := validateImageFiles(rows); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	c, err := readConnection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	counts, err := applySeed(c, rows)
	if err != nil {
		message := err.Error()
		if c.Password != "" {
			message = strings.ReplaceAll(message, c.Password, "[REDACTED]")
		}
		fmt.Fprintln(os.Stderr, "Seed failed; no application-row updates/deletes attempted:", message)
		os.Exit(1)
	}
	fmt.Printf("Seed committed to LOCAL %s. Created managers=%d categories=%d restaurants=%d menu_items=%d; reused seed records=%d.\n", databaseName, counts.Managers, counts.Categories, counts.Restaurants, counts.Items, counts.Reused)
}
func applySeed(c connection, rows []restaurant) (stats, error) {
	var counts stats
	if err := validateConnection(c); err != nil {
		return counts, err
	}
	db, err := gorm.Open(postgres.Open(c.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), TranslateError: true})
	if err != nil {
		return counts, errors.New("cannot connect to approved local database; connection details suppressed")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return counts, err
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target struct {
			Name    string
			Address string
		}
		if err := tx.Raw("SELECT current_database() AS name, host(inet_server_addr()) AS address").Scan(&target).Error; err != nil {
			return err
		}
		if target.Name != databaseName || (target.Address != "127.0.0.1" && target.Address != "::1") {
			return errors.New("connected target is not the approved loopback database")
		}
		if err := tx.Exec("SELECT pg_advisory_xact_lock(20261101,2)").Error; err != nil {
			return err
		}
		if err := tx.Exec("LOCK TABLE public.users, public.categories, public.restaurants, public.menu_items IN SHARE ROW EXCLUSIVE MODE").Error; err != nil {
			return err
		}
		var unsafe int64
		if err := tx.Raw(`SELECT count(*) FROM pg_trigger WHERE tgrelid IN ('public.users'::regclass,'public.categories'::regclass,'public.restaurants'::regclass,'public.menu_items'::regclass) AND NOT tgisinternal`).Scan(&unsafe).Error; err != nil {
			return err
		}
		if unsafe != 0 {
			return errors.New("custom triggers require review before seeding")
		}
		if err := tx.Raw(`SELECT count(*) FROM pg_rewrite WHERE ev_class IN ('public.users'::regclass,'public.categories'::regclass,'public.restaurants'::regclass,'public.menu_items'::regclass) AND rulename <> '_RETURN'`).Scan(&unsafe).Error; err != nil {
			return err
		}
		if unsafe != 0 {
			return errors.New("custom insert/rewrite rules require review before seeding")
		}
		if err := tx.Raw(`SELECT count(*) FROM pg_class WHERE oid IN ('public.users'::regclass,'public.categories'::regclass,'public.restaurants'::regclass,'public.menu_items'::regclass) AND (relkind <> 'r' OR relrowsecurity OR relispartition)`).Scan(&unsafe).Error; err != nil {
			return err
		}
		if unsafe != 0 {
			return errors.New("unexpected table kind/RLS/partition requires review")
		}
		var foreignKeys int64
		if err := tx.Raw(`SELECT count(*) FROM pg_constraint c JOIN pg_namespace n ON n.oid=c.connamespace WHERE n.nspname='public' AND c.contype='f' AND (
   (c.conrelid='public.restaurants'::regclass AND c.confrelid='public.users'::regclass AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid AND attname='manager_id')]::smallint[] AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.confrelid AND attname='id')]::smallint[]) OR
   (c.conrelid='public.menu_items'::regclass AND c.confrelid='public.restaurants'::regclass AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid AND attname='restaurant_id')]::smallint[] AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.confrelid AND attname='id')]::smallint[]) OR
   (c.conrelid='public.menu_items'::regclass AND c.confrelid='public.categories'::regclass AND c.conkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.conrelid AND attname='category_id')]::smallint[] AND c.confkey=ARRAY[(SELECT attnum FROM pg_attribute WHERE attrelid=c.confrelid AND attname='id')]::smallint[]))`).Scan(&foreignKeys).Error; err != nil {
			return err
		}
		if foreignKeys != 3 {
			return errors.New("expected manager/restaurant/category foreign keys missing or ambiguous")
		}
		if err := ensureRegistry(tx); err != nil {
			return err
		}
		for _, r := range rows {
			managerID, err := ensureManager(tx, r, &counts)
			if err != nil {
				return err
			}
			restaurantID, err := ensureRestaurant(tx, r, managerID, &counts)
			if err != nil {
				return err
			}
			for _, d := range r.Items {
				categoryID, err := ensureCategory(tx, d.Category, &counts)
				if err != nil {
					return err
				}
				if err := ensureDish(tx, r.Key, d, restaurantID, categoryID, &counts); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return counts, err
}
func ensureRegistry(tx *gorm.DB) error {
	var namespace string
	if err := tx.Raw("SELECT coalesce(obj_description(to_regnamespace('foodhub_demo_seed'),'pg_namespace'),'')").Scan(&namespace).Error; err != nil {
		return err
	}
	var exists bool
	if err := tx.Raw("SELECT to_regnamespace('foodhub_demo_seed') IS NOT NULL").Scan(&exists).Error; err != nil {
		return err
	}
	if exists && namespace != seedMarker {
		return errors.New("unrecognised foodhub_demo_seed namespace; refusing to adopt existing objects")
	}
	if !exists {
		if err := tx.Exec("CREATE SCHEMA foodhub_demo_seed").Error; err != nil {
			return err
		}
		if err := tx.Exec("COMMENT ON SCHEMA foodhub_demo_seed IS 'foodhub-local-presentation-seed:v1'").Error; err != nil {
			return err
		}
		return tx.Exec(`CREATE TABLE foodhub_demo_seed.records (
   kind text NOT NULL CHECK(kind IN ('manager','restaurant','menu_item')),
   seed_key text NOT NULL, entity_id bigint NOT NULL CHECK(entity_id > 0),
   PRIMARY KEY(kind,seed_key), UNIQUE(kind,entity_id))`).Error
	}
	// Existing registry must be readable; never repair or recreate it automatically.
	var count int64
	return tx.Table("foodhub_demo_seed.records").Count(&count).Error
}
func recordedID(tx *gorm.DB, kind, key string) (uint, error) {
	var row record
	err := tx.Table("foodhub_demo_seed.records").Where("kind=? AND seed_key=?", kind, key).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return row.EntityID, err
}
func remember(tx *gorm.DB, kind, key string, id uint) error {
	return tx.Table("foodhub_demo_seed.records").Create(&record{kind, key, id}).Error
}
func ensureManager(tx *gorm.DB, r restaurant, s *stats) (uint, error) {
	email := r.Key + ".manager@foodhub-demo.example.test"
	id, err := recordedID(tx, "manager", r.Key)
	if err != nil {
		return 0, err
	}
	if id != 0 {
		var user models.User
		if err := tx.Unscoped().First(&user, id).Error; err != nil {
			return 0, err
		}
		if user.DeletedAt.Valid || user.Email != email || user.Role != models.RoleManager || user.Status != models.UserStatusActive {
			return 0, fmt.Errorf("seed manager drift: %s (no changes made)", r.Key)
		}
		s.Reused++
		return id, nil
	}
	var count int64
	if err := tx.Unscoped().Model(&models.User{}).Where("lower(trim(email))=lower(?)", email).Count(&count).Error; err != nil {
		return 0, err
	}
	if count != 0 {
		return 0, fmt.Errorf("existing email collision: %s; refusing to adopt or modify account", r.Key)
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return 0, err
	}
	hash, err := utils.HashPassword(hex.EncodeToString(secret))
	if err != nil {
		return 0, err
	}
	user := models.User{FirstName: "Demo", LastName: r.Name, Email: email, Phone: r.Phone, Password: hash, Role: models.RoleManager, Status: models.UserStatusActive}
	if err := tx.Omit(clause.Associations).Create(&user).Error; err != nil {
		return 0, err
	}
	if err := remember(tx, "manager", r.Key, user.ID); err != nil {
		return 0, err
	}
	s.Managers++
	return user.ID, nil
}
func ensureRestaurant(tx *gorm.DB, r restaurant, managerID uint, s *stats) (uint, error) {
	id, err := recordedID(tx, "restaurant", r.Key)
	if err != nil {
		return 0, err
	}
	email := r.Key + "@foodhub-demo.example.test"
	var count int64
	if err := tx.Unscoped().Model(&models.Restaurant{}).Where("(lower(trim(name))=lower(?) OR lower(trim(email))=lower(?)) AND id<>?", r.Name, email, id).Count(&count).Error; err != nil {
		return 0, err
	}
	if count != 0 {
		return 0, fmt.Errorf("existing restaurant collision: %s; refusing to adopt or modify it", r.Name)
	}
	if id != 0 {
		var existing models.Restaurant
		if err := tx.Unscoped().First(&existing, id).Error; err != nil {
			return 0, err
		}
		if existing.DeletedAt.Valid || existing.ManagerID != managerID || existing.Name != r.Name || existing.Description != r.Description || existing.Email != email || existing.Status != models.RestaurantStatusActive || existing.Address != r.Address || existing.City != r.City || existing.State != r.State || existing.Phone != r.Phone || existing.OpeningTime != r.OpeningTime || existing.ClosingTime != r.ClosingTime || existing.ImageURL != r.ImageURL {
			return 0, fmt.Errorf("seed restaurant drift: %s; refusing to update", r.Key)
		}
		s.Reused++
		return id, nil
	}
	row := restaurantRecord(r, managerID)
	if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
		return 0, err
	}
	if err := remember(tx, "restaurant", r.Key, row.ID); err != nil {
		return 0, err
	}
	s.Restaurants++
	return row.ID, nil
}
func ensureCategory(tx *gorm.DB, name string, s *stats) (uint, error) {
	var matches []models.Category
	if err := tx.Unscoped().Where("lower(trim(name))=lower(?)", name).Find(&matches).Error; err != nil {
		return 0, err
	}
	if len(matches) > 1 || (len(matches) == 1 && matches[0].DeletedAt.Valid) {
		return 0, fmt.Errorf("ambiguous/deleted category %s; refusing to change it", name)
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}
	row := models.Category{Name: name}
	if err := tx.Omit(clause.Associations).Create(&row).Error; err != nil {
		return 0, err
	}
	s.Categories++
	return row.ID, nil
}
func ensureDish(tx *gorm.DB, restaurantKey string, d dish, restaurantID, categoryID uint, s *stats) error {
	key := restaurantKey + "/" + d.Key
	id, err := recordedID(tx, "menu_item", key)
	if err != nil {
		return err
	}
	var count int64
	if err := tx.Unscoped().Model(&models.MenuItem{}).Where("restaurant_id=? AND lower(trim(name))=lower(?) AND id<>?", restaurantID, d.Name, id).Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("existing menu collision: %s; refusing to adopt/modify", key)
	}
	if id != 0 {
		var existing models.MenuItem
		if err := tx.Unscoped().First(&existing, id).Error; err != nil {
			return err
		}
		if existing.DeletedAt.Valid || existing.RestaurantID != restaurantID || existing.CategoryID != categoryID || existing.Name != d.Name || existing.Description != d.Description || existing.Price != d.Price || existing.Available != d.Available || existing.ImageURL != d.ImageURL {
			return fmt.Errorf("seed menu drift: %s; refusing to update", key)
		}
		s.Reused++
		return nil
	}
	// Explicit SQL preserves available=false; GORM's default:true would replace a
	// false struct value on Create. RETURNING supplies the real generated FK ID.
	row := menuRecord(d, restaurantID, categoryID)
	var inserted struct{ ID uint }
	if err := tx.Raw(`INSERT INTO public.menu_items(created_at,updated_at,restaurant_id,category_id,name,description,price,image_url,available)
  VALUES(CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?,?,?,?,?,?,?) RETURNING id`, row.RestaurantID, row.CategoryID, row.Name, row.Description, row.Price, row.ImageURL, row.Available).Scan(&inserted).Error; err != nil {
		return err
	}
	if inserted.ID == 0 {
		return errors.New("menu insert returned no ID")
	}
	if err := remember(tx, "menu_item", key, inserted.ID); err != nil {
		return err
	}
	s.Items++
	return nil
}

// These same mappings are exercised by offline cart/checkout contract tests.
func restaurantRecord(r restaurant, managerID uint) models.Restaurant {
	return models.Restaurant{ManagerID: managerID, Name: r.Name, Description: r.Description, Address: r.Address, City: r.City, State: r.State, Phone: r.Phone, Email: r.Key + "@foodhub-demo.example.test", OpeningTime: r.OpeningTime, ClosingTime: r.ClosingTime, ImageURL: r.ImageURL, Status: models.RestaurantStatusActive}
}
func menuRecord(d dish, restaurantID, categoryID uint) models.MenuItem {
	return models.MenuItem{RestaurantID: restaurantID, CategoryID: categoryID, Name: d.Name, Description: d.Description, Price: d.Price, ImageURL: d.ImageURL, Available: d.Available}
}
