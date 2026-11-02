-- Read-only readiness check. Does not execute the seed or advance any sequences.
BEGIN READ ONLY;
SET LOCAL statement_timeout = '30s';
SELECT current_database() AS database, current_user AS role, host(inet_server_addr()) AS server,
       has_database_privilege(current_user,current_database(),'CREATE') AS can_create_seed_schema,
       has_schema_privilege(current_user,'public','USAGE') AS public_usage;
SELECT c.oid::regclass AS relation, pg_get_userbyid(c.relowner) AS owner,
       has_table_privilege(current_user,c.oid,'SELECT') AS can_select,
       has_table_privilege(current_user,c.oid,'INSERT') AS can_insert,
       has_table_privilege(current_user,c.oid,'UPDATE') AS can_acquire_write_lock,
       c.relkind, c.relrowsecurity, c.relispartition
FROM pg_class c WHERE c.oid IN ('public.users'::regclass,'public.categories'::regclass,
 'public.restaurants'::regclass,'public.menu_items'::regclass);
SELECT conrelid::regclass AS relation,conname,pg_get_constraintdef(oid) AS definition
FROM pg_constraint WHERE contype='f' AND conrelid IN ('public.restaurants'::regclass,'public.menu_items'::regclass);
SELECT count(*) AS custom_triggers FROM pg_trigger
WHERE tgrelid IN ('public.users'::regclass,'public.categories'::regclass,'public.restaurants'::regclass,'public.menu_items'::regclass)
AND NOT tgisinternal;
SELECT count(*) AS custom_rewrite_rules FROM pg_rewrite
WHERE ev_class IN ('public.users'::regclass,'public.categories'::regclass,'public.restaurants'::regclass,'public.menu_items'::regclass)
AND rulename <> '_RETURN';
SELECT to_regnamespace('foodhub_demo_seed') IS NOT NULL AS registry_schema_exists,
       obj_description(to_regnamespace('foodhub_demo_seed'),'pg_namespace') AS registry_marker;
WITH expected(name,email,manager_email) AS (VALUES
 ('Tokyo Bento Kitchen','tokyo-bento-kitchen@foodhub-demo.example.test','tokyo-bento-kitchen.manager@foodhub-demo.example.test'),
 ('Burger District','burger-district@foodhub-demo.example.test','burger-district.manager@foodhub-demo.example.test'),
 ('Bella Italia Kitchen','bella-italia-kitchen@foodhub-demo.example.test','bella-italia-kitchen.manager@foodhub-demo.example.test'),
 ('Spice Route','spice-route@foodhub-demo.example.test','spice-route.manager@foodhub-demo.example.test'),
 ('The Shawarma House','shawarma-house@foodhub-demo.example.test','shawarma-house.manager@foodhub-demo.example.test'),
 ('Sweet Treats Café','sweet-treats-cafe@foodhub-demo.example.test','sweet-treats-cafe.manager@foodhub-demo.example.test')
)
SELECT e.name,
 (SELECT count(*) FROM public.restaurants r WHERE lower(trim(r.name))=lower(e.name) OR lower(trim(r.email))=lower(e.email)) AS restaurant_collisions,
 (SELECT count(*) FROM public.users u WHERE lower(trim(u.email))=lower(e.manager_email)) AS manager_email_collisions
FROM expected e;
WITH expected(name) AS (VALUES ('Burgers'),('Sides'),('Wings'),('Pizzas'),('Pasta'),('Biryani'),('Curries'),('Breads'),('Wraps'),('Grills'),('Waffles'),('Cakes'),('Milkshakes'),('Japanese Rice Bowls'),('Sushi Rolls'),('Japanese Noodles'),('Japanese Sides'),('Japanese Desserts'))
SELECT e.name,count(c.id) AS existing_matches,count(c.id) FILTER(WHERE c.deleted_at IS NOT NULL) AS soft_deleted_matches
FROM expected e LEFT JOIN public.categories c ON lower(trim(c.name))=lower(e.name)
GROUP BY e.name ORDER BY e.name;
-- Aggregate baselines only: no customer identity/passwords are printed.
SELECT (SELECT count(*) FROM public.users) AS users,
       (SELECT count(*) FROM public.restaurants) AS restaurants,
       (SELECT count(*) FROM public.menu_items) AS menu_items,
       (SELECT count(*) FROM public.categories) AS categories,
       (SELECT count(*) FROM public.carts) AS carts,
       (SELECT count(*) FROM public.orders) AS orders;
SELECT schemaname,sequencename,last_value,increment_by,
       has_sequence_privilege(current_user,format('%I.%I',schemaname,sequencename),'USAGE') AS can_use_sequence
FROM pg_sequences WHERE schemaname='public' AND sequencename IN ('users_id_seq','restaurants_id_seq','menu_items_id_seq','categories_id_seq');
SELECT 'users' AS relation,max(id) AS greatest_id FROM public.users
UNION ALL SELECT 'restaurants',max(id) FROM public.restaurants
UNION ALL SELECT 'menu_items',max(id) FROM public.menu_items
UNION ALL SELECT 'categories',max(id) FROM public.categories;
COMMIT;
