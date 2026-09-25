alter table app_metrics
  alter column crashes type numeric(6,3) using crashes::numeric,
  alter column anr_count type numeric(6,3) using anr_count::numeric;

alter table app_metrics rename column crashes to crash_rate;
alter table app_metrics rename column anr_count to anr_rate;
