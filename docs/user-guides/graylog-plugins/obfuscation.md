The Graylog Obfuscation Plugin scans input logs, searches text sensitive data, and anonymizes it.
The search operation is performed using a regular expression.
The plugin support filtration by streams and fields. Only selected streams and fields are obfuscated.

The plugin configuration is displayed in the **Configurations** page as shown in the image below.

![Graylog Obfuscation Plugin Configuration](/docs/images/plugins/obfuscation-configuration.png)

# Configuration Parameters

The **Obfuscation Plugin Configuration** page with its parameters is shown in the image below.

![Obfuscation Configuration Page](/docs/images/plugins/obfuscation-plugin-configuration-page.png)

The plugin configuration parameters are as follows:

* `Is Obfuscation Enabled` - If checked, then the obfuscation plugin is enabled.
* `Text Replacer` - The pattern used to replace sensitive data in the text. The following text replacer is supported:
  * Static star replacer, for example, ********.
* `Field Names` - The field names in the log message that are to be processed by the obfuscation plugin.
  The obfuscation is supported only for text fields. The field names should be unique.
* `Stream Titles` - The messages from the streams that are to be processed by the obfuscation plugin.
  The stream titles should be unique. The stream filter will work only if Pipeline Processor and Message Filter Chain
  is enabled and located higher, than Message Obfuscator.

![Message Processors Configuration](/docs/images/plugins/message-processors-configuration.png)

* `Sensitive Regular Expressions` - The regex used for catching sensitive data in the text. For example, `find`.
  This parameter includes the following information:
  * ID - The unique ID of the regex.
  * Name - The readable name.
  * Pattern - The regex pattern written for the java regex engine.
  * Importance - The number used to resolve conflicts in the search. If two regex catch different parts in one word,
    the conflict is resolved by using the maximum importance value of the regex patterns.
* `White Regular Expressions` - The regex used for filtering white words from all the catch sensitive data.
  For example, `matched`. This parameter includes the following information:
  * ID - The unique ID of the regex.
  * Name - The readable name.
  * Pattern - The regex pattern written for the java regex engine.

## Buttons

* The **Reset** button resets the obfuscation configuration to its default.
* The **Save** button stores the obfuscation configuration. The configuration is applied immediately.
