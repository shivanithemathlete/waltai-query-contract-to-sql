import duckdb

db = duckdb.connect(":memory:")

db.execute("""
CREATE TABLE dim_store (
    store_id INTEGER,
    region VARCHAR
);

CREATE TABLE dim_calendar (
    date DATE,
    fiscal_year INTEGER
);

CREATE TABLE fact_sales (
    store_id INTEGER,
    order_date DATE,
    order_id VARCHAR,
    channel VARCHAR,
    revenue DOUBLE
);

INSERT INTO dim_store VALUES
(1, 'North'), (2, 'South'), (3, 'West');

INSERT INTO dim_calendar VALUES
('2025-01-01', 2025),
('2026-01-01', 2026);

INSERT INTO fact_sales VALUES
(1, '2025-01-01', 'O-201', 'Offline', 100),
(1, '2026-01-01', 'O-202', 'Online', 100),
(1, '2026-01-01', 'O-203', 'Online', 100),
(1, '2026-01-01', 'O-204', 'Online', 50),
(2, '2025-01-01', 'O-205', 'Offline', 200),
(2, '2026-01-01', 'O-203', 'Online', 170),
(2, '2026-01-01', 'O-206', 'Offline', 200),
(3, '2025-01-01', 'O-207', 'Offline', 150),
(3, '2026-01-01', 'O-208', 'Offline', 90);
""")

# Paste your generated SQL here.
# a
query_a = """
SELECT
    SUM(CASE WHEN t0.channel = 'Online' THEN t0.revenue END) AS online_rev,
    SUM(t0.revenue) AS total_revenue
FROM fact_sales t0
INNER JOIN dim_calendar t1 ON t0.order_date = t1.date
WHERE t1.fiscal_year = 2026
"""

# b
query_b = """
SELECT
    t1.region AS region,
SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END) AS total_revenue_2025,
    SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) AS total_revenue_2026,
    (SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) - SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END)) AS total_revenue_delta,
    (SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) - SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END)) * 100.0 / NULLIF(SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END), 0) AS total_revenue_pct_change
FROM fact_sales t0
LEFT JOIN dim_store t1 ON t0.store_id = t1.store_id
LEFT JOIN dim_calendar t2 ON t0.order_date = t2.date
GROUP BY t1.region
UNION ALL
SELECT
    'Total' AS region,
SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END) AS total_revenue_2025,
    SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) AS total_revenue_2026,
    (SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) - SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END)) AS total_revenue_delta,
    (SUM(CASE WHEN t2.fiscal_year = '2026' THEN t0.revenue END) - SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END)) * 100.0 / NULLIF(SUM(CASE WHEN t2.fiscal_year = '2025' THEN t0.revenue END), 0) AS total_revenue_pct_change
FROM fact_sales t0
LEFT JOIN dim_store t1 ON t0.store_id = t1.store_id
LEFT JOIN dim_calendar t2 ON t0.order_date = t2.date
"""

query_c = """
SELECT
    t1.region AS region,
    COUNT(DISTINCT t0.order_id) AS order_count
FROM fact_sales t0
LEFT JOIN dim_store t1 ON t0.store_id = t1.store_id
GROUP BY t1.region
UNION ALL
SELECT
    'Total' AS region,
    COUNT(DISTINCT t0.order_id) AS order_count
FROM fact_sales t0
LEFT JOIN dim_store t1 ON t0.store_id = t1.store_id
"""

try:
    result = db.execute(query_a)
    print("Columns:", [col[0] for col in result.description])
    for row in result.fetchall():
        print(row)
except Exception as e:
    print("SQL FAILED:", e)