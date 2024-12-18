-- different test strings
local test_strings = {
    -- correct: valid syslog levels, from 0 to 7
    "emerg",
    "alert",
    "crit",
    "err",
    "warning",
    "notice",
    "info",
    "debug",
    -- correct: full level names
    "emergency",
    "alert",
    "critical",
    "error",
    "warning",
    "notice",
    "info",
    "debug",
    -- correct: levels using capital letters
    "EMERG",
    "ALERT",
    "CRTI",
    "ERR",
    "WARNING",
    "NOTICE",
    "INFO",
    "DEBUG",
    -- correct: full level names using upper case
    "EMERGENCY",
    "ALERT",
    "CRITICAL",
    "ERROR",
    "WARNING",
    "NOTICE",
    "INFO",
    "DEBUG",
    -- correct: other short or full level names forms
    "warn",
    "fatal",
    "trace",
    -- incorrect: short level names
    "emg",
    "alrt",
    "art",
    "alt",
    "crt",
    "wrg",
    "wrn",
    "inf",
    "dbg",
    -- incorrect: various combinations of levels
    "er",
    "E",
    "wa",
    "war",
    "W",
    "ntc",
    "noti",
    "N",
    "in",
    "I",
    "deb",
    "D",
    "fat",
    "F",
    -- incorrect: words which were parsed as levels
    "number",
}

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

-- test functions
-- call "like real" functions
function execute_real_func_test()
    for i, test_string in ipairs(test_strings) do
        local test_structure = {}
        test_structure["level"] = test_string

        local start_time = os.time()
        local code, time, new_test_structure = update_level("test", i, test_structure)
        local end_time = os.time()

        for k,v in pairs(new_test_structure) do
            if (k == "level" and code == 2) then
                print("Original string:", test_string)
                print("Call kv_parse = ", start_time)
                print("Complete update_level = ", end_time, "Execution time =", end_time - start_time)
                print("Code:", code, "Processing order:", time)
                print("Update level:", test_string, "=>", v)
                print("------------------------------------------------------------------------")
            end
            if (k == "level" and code == 0) then
                -- although this level can be ignore by script, but this level will validate by regex
                print ("Level was ignored by script:", test_string)
            end
        end
    end
end

print("====================================================================")
print("Run test to check function which will use Fluent")
print("====================================================================")
execute_real_func_test()
