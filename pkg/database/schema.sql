-- Create table for User
CREATE TABLE User (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    phone_number TEXT,
    user_type TEXT,
    profile TEXT,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE HorseRaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    horse_name TEXT NOT NULL,
    event TEXT NOT NULL,
    price TEXT NOT NULL,
    selection_id INTEGER NOT NULL,
    event_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Create table for MarketData
CREATE TABLE MarketData (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    menu_hint TEXT,
    event_name TEXT,
    event_dt TEXT,
    selection_id INTEGER,
    selection_name TEXT,
    win_lose TEXT,
    bsp REAL,
    ppwap REAL,
    morning_wap REAL,
    ppmax REAL,
    ppmin REAL,
    ipmax REAL,
    ipmin REAL,
    morning_traded_vol REAL,
    pp_traded_vol REAL,
    ip_traded_vol REAL,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Configurations (
    ID    INTEGER PRIMARY KEY AUTOINCREMENT,
    key   TEXT    UNIQUE
                  NOT NULL,
    value TEXT    NOT NULL
);


-- Create table for Selection --
-- Select relevant data for the given horse
WITH horse_data AS (
    SELECT
        win_lose,
        ipmin,
        ppmin,
        morning_wap AS morning_ppmax,
        ipmax,
        pp_traded_vol,
        selection_name,
        selection_id
    FROM MarketData
    WHERE selection_id = 13415380
),
stats AS (
    SELECT
        AVG(win_lose) AS avg_win_lose,
        AVG(ipmin) AS avg_ipmin,
        AVG(ppmin) AS avg_ppmin,
        AVG(morning_ppmax) AS avg_morning_ppmax,
        AVG(ipmax) AS avg_ipmax,
        AVG(pp_traded_vol) AS avg_pp_traded_vol,
        COUNT(*) AS n
    FROM horse_data
),
correlations AS (
    SELECT
        selection_name,
        selection_id,
        (
            SUM((win_lose - avg_win_lose) * (ipmin - avg_ipmin)) / (n - 1)
        ) / 
        (
            (SQRT(SUM(POW((win_lose - avg_win_lose), 2)) / (n - 1))) *
            (SQRT(SUM(POW((ipmin - avg_ipmin), 2)) / (n - 1)))
        ) AS win_lose_ipmin,

        (
            SUM((win_lose - avg_win_lose) * (ppmin - avg_ppmin)) / (n - 1)
        ) / 
        (
            (SQRT(SUM(POW((win_lose - avg_win_lose), 2)) / (n - 1))) *
            (SQRT(SUM(POW((ppmin - avg_ppmin), 2)) / (n - 1)))
        ) AS win_lose_ppmin,

        (
            SUM((win_lose - avg_win_lose) * (morning_ppmax - avg_morning_ppmax)) / (n - 1)
        ) / 
        (
            (SQRT(SUM(POW((win_lose - avg_win_lose), 2)) / (n - 1))) *
            (SQRT(SUM(POW((morning_ppmax - avg_morning_ppmax), 2)) / (n - 1)))
        ) AS win_lose_morning_ppmax,

        (
            SUM((win_lose - avg_win_lose) * (ipmax - avg_ipmax)) / (n - 1)
        ) / 
        (
            (SQRT(SUM(POW((win_lose - avg_win_lose), 2)) / (n - 1))) *
            (SQRT(SUM(POW((ipmax - avg_ipmax), 2)) / (n - 1)))
        ) AS win_lose_ipmax,

        (
            SUM((win_lose - avg_win_lose) * (pp_traded_vol - avg_pp_traded_vol)) / (n - 1)
        ) / 
        (
            (SQRT(SUM(POW((win_lose - avg_win_lose), 2)) / (n - 1))) *
            (SQRT(SUM(POW((pp_traded_vol - avg_pp_traded_vol), 2)) / (n - 1)))
        ) AS win_lose_pptraded
    FROM horse_data, stats
)
SELECT * FROM correlations;




-- Correlation Analysis
-- WIN_LOSE has strong negative correlations with IPMIN (-1.0), PPMIN (-0.67), and MORNING_PPMAX (-0.61).
-- WIN_LOSE has a positive correlation with IPMAX (0.39) and PPTRADED (0.37).
-- Interpretation
-- Lower MORNING_PPMAX, PPMIN, and IPMIN values might indicate a higher chance of winning.
-- Higher IPMAX values and PPTRADED amounts might also correlate with better performance.
