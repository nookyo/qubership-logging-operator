-- Fluent Bit supports only next levels:
-- "emerg", "alert", "crit", "err", "warning", "notice", "info", "debug"
-- Fluent Bit source code of gelf output:
-- https://github.com/fluent/fluent-bit/blob/master/src/flb_pack_gelf.c#L563-L592
-- this script marks non supported levels with syslog codes
-- input: https://docs.fluentbit.io/manual/pipeline/filters/lua#function-arguments
-- output: https://docs.fluentbit.io/manual/pipeline/filters/lua#return-values
function update_level(tag, timestamp, record)
  if (record["level"] ~= nil) then
    record["level"] = string.lower(record["level"])

    -- return 2 everywhere inside the if block,
    -- that means the original timestamp is not modified and the record has been modified
    if (record["level"] == "error") then
        record["level"] = "err"
        return 2, timestamp, record
    end
    if (record["level"] == "warn") then
        record["level"] = "warning"
        return 2, timestamp, record
    end
    if (record["level"] == "trace") then
      record["level"] = "debug"
      return 2, timestamp, record
    end
    if (record["level"] == "fatal") then
      record["level"] = "emerg"
      return 2, timestamp, record
    end
    if (record["level"] == "critical") then
      record["level"] = "crit"
      return 2, timestamp, record
    end

    -- return 0, that the record will not be modified
    return 0, timestamp, record
  end

  -- return 0, that the record will not be modified
  return 0, timestamp, record
end