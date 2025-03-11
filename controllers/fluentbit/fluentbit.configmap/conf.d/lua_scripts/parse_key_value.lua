-- input: https://docs.fluentbit.io/manual/pipeline/filters/lua#function-arguments
-- output: https://docs.fluentbit.io/manual/pipeline/filters/lua#return-values
function kv_parse(tag, timestamp, record)
    if record["log"] ~= nil and type(record["log"]) ~= "table" then
        -- regex to find the end of key=value string in the original string
        -- this regex search the place:
        -- * start from ]
        -- * with 0 or more space symbols
        -- * without [
        -- * start from alphabet symbol, digit or any symbol (expect [)
        local regex_kvs_end = "]%s*[^%[][%w%-%{%}%\\%/%.%,%!%@%#%$%%%^%&%*%(%)]%s*"
        local regex_kvs = "%[([^=%[%]]+)=(%w*(.[^%[^%]]*))%]"
        local s = record["log"]

        -- find the end position of [key=value] pairs
        -- and copy from original string only this string part, for example:
        -- [<time>] [INFO] [key1=value1][key2=value2] ... [keyN=valueN]
        local kvs_position = string.find(s, regex_kvs_end, 1)
        local kvs = string.sub(s, 0, kvs_position)

        if kvs ~= nil then
            for k, v in string.gmatch(kvs, regex_kvs) do
              record[k] = v
            end
        else
            -- return 0, that means the record will not be modified
            return 0, timestamp, record
        end

        -- return 2, that means the original timestamp is not modified and the record has been modified
        -- so it must be replaced by the returned values from the record
        return 2, timestamp, record
    else
        -- return 0, that means the record will not be modified
        return 0, timestamp, record
    end
end
