alter table app_metrics rename column crash_rate to crashes;
alter table app_metrics rename column anr_rate to anr_count;

alter table app_metrics
  alter column crashes type integer using round(crashes)::integer,
  alter column anr_count type integer using round(anr_count)::integer;
