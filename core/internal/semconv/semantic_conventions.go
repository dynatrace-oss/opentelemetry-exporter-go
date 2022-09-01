// Copyright 2022 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semconv

// Prefixes for semantic conventions
const (
	// Prefix for 'cloud'
	CloudPrefix = "cloud"

	// Prefix for 'container'
	ContainerPrefix = "container"

	// Prefix for 'dt.cloud.gcp'
	DtCloudGcpPrefix = "gcp"

	// Prefix for 'dt.ctg_resource'
	DtCtgResourcePrefix = "dt.ctg"

	// Prefix for 'dt.env_vars'
	DtEnvVarsPrefix = "dt.env_vars"

	// Prefix for 'dt.ims_resource'
	DtImsResourcePrefix = "dt.ims"

	// Prefix for 'dt.os'
	DtOsPrefix = "dt.os"

	// Prefix for 'dt.osi'
	DtOsiPrefix = "dt.host"

	// Prefix for 'dt.pgi'
	DtPgiPrefix = "dt.process"

	// Prefix for 'dt.tech'
	DtTechPrefix = "dt.tech"

	// Prefix for 'dt.telemetry.exporter'
	DtTelemetryExporterPrefix = "telemetry.exporter"

	// Prefix for 'dt.websphere'
	DtWebspherePrefix = "dt.websphere"

	// Prefix for 'dt.zosconnect_resource'
	DtZosconnectResourcePrefix = "dt.zosconnect"

	// Prefix for 'faas_resource'
	FaasResourcePrefix = "faas"

	// Prefix for 'host'
	HostPrefix = "host"

	// Prefix for 'process'
	ProcessPrefix = "process"

	// Prefix for 'process.executable'
	ProcessExecutablePrefix = "process.executable"

	// Prefix for 'process.runtime'
	ProcessRuntimePrefix = "process.runtime"

	// Prefix for 'service'
	ServicePrefix = "service"

	// Prefix for 'telemetry.sdk'
	TelemetrySdkPrefix = "telemetry.sdk"

	// Prefix for 'aws.lambda'
	AwsLambdaPrefix = "aws.lambda"

	// Prefix for 'code'
	CodePrefix = "code"

	// Prefix for 'db'
	DbPrefix = "db"

	// Prefix for 'db.mssql'
	DbMssqlPrefix = "db.mssql"

	// Prefix for 'db.cassandra'
	DbCassandraPrefix = "db.cassandra"

	// Prefix for 'db.hbase'
	DbHbasePrefix = "db.hbase"

	// Prefix for 'db.redis'
	DbRedisPrefix = "db.redis"

	// Prefix for 'db.mongodb'
	DbMongodbPrefix = "db.mongodb"

	// Prefix for 'dt.db'
	DtDbPrefix = "dt.db"

	// Prefix for 'dt.exception'
	DtExceptionPrefix = "dt.exception"

	// Prefix for 'dt.faas'
	DtFaasPrefix = "dt.faas"

	// Prefix for 'dt.code'
	DtCodePrefix = "dt.code"

	// Prefix for 'dt.stacktrace'
	DtStacktracePrefix = "dt.stacktrace"

	// Prefix for 'dt.http.server'
	DtHttpServerPrefix = "dt.http"

	// Prefix for 'otel.library'
	OtelLibraryPrefix = "otel.library"

	// Prefix for 'dt.messaging'
	DtMessagingPrefix = "dt.messaging"

	// Prefix for 'dt.parent'
	DtParentPrefix = "dt.parent"

	// Prefix for 'dt.rum'
	DtRumPrefix = "dt.rum"

	// Prefix for 'faas_span.datasource'
	FaasSpanDatasourcePrefix = "faas.document"

	// Prefix for 'network'
	NetworkPrefix = "net"

	// Prefix for 'identity'
	IdentityPrefix = "enduser"

	// Prefix for 'http'
	HttpPrefix = "http"

	// Prefix for 'aws'
	AwsPrefix = "aws"

	// Prefix for 'dynamodb.shared'
	DynamodbSharedPrefix = "aws.dynamodb"

	// Prefix for 'messaging'
	MessagingPrefix = "messaging"

	// Prefix for 'rpc'
	RpcPrefix = "rpc"

	// Prefix for 'rpc.grpc'
	RpcGrpcPrefix = "rpc.grpc"

	// Prefix for 'rpc.jsonrpc'
	RpcJsonrpcPrefix = "rpc.jsonrpc"

	// Prefix for 'rpc.message'
	RpcMessagePrefix = "message"
)

// Attribute definitions
const (
	// Name of the cloud provider.
	// This attribute expects a value of type string from the enumeration CloudProviderValues.
	CloudProvider = "cloud.provider"

	// The cloud account ID the resource is assigned to.
	// This attribute expects a value of type string.
	CloudAccountId = "cloud.account.id"

	// The geographical region the resource is running. Refer to your provider's docs to see the available regions, for example [AWS regions](https://aws.amazon.com/about-aws/global-infrastructure/regions_az/), [Azure regions](https://azure.microsoft.com/en-us/global-infrastructure/geographies/), or [Google Cloud regions](https://cloud.google.com/about/locations).
	// This attribute expects a value of type string.
	CloudRegion = "cloud.region"

	// Cloud regions often have multiple, isolated locations known as zones to increase availability. Availability zone represents the zone where the resource is running.
	// This attribute expects a value of type string.
	//
	// Availability zones are called "zones" on Google Cloud.
	CloudAvailabilityZone = "cloud.availability_zone"

	// The cloud platform in use.
	// This attribute expects a value of type string from the enumeration CloudPlatformValues.
	//
	// The prefix of the service SHOULD match the one specified in `cloud.provider`.
	CloudPlatform = "cloud.platform"

	// Container name.
	// This attribute expects a value of type string.
	ContainerName = "container.name"

	// Name of the image the container was built on.
	// This attribute expects a value of type string.
	ContainerImageName = "container.image.name"

	// Container image tag.
	// This attribute expects a value of type string.
	ContainerImageTag = "container.image.tag"

	// A project organizes all your Google Cloud resources.
	// This attribute expects a value of type string.
	GcpProjectId = "gcp.project.id"

	// A region is a specific geographical location where you can host your resources.
	// This attribute expects a value of type string.
	GcpRegion = "gcp.region"

	// A permanent identifier that is unique within your Google Cloud project.
	// This attribute expects a value of type string.
	GcpInstanceId = "gcp.instance.id"

	// The name to display for the instance in the Cloud Console.
	// This attribute expects a value of type string.
	GcpInstanceName = "gcp.instance.name"

	// The name of a resource type.
	// This attribute expects a value of type string.
	GcpResourceType = "gcp.resource.type"

	// Deprecated: Use `dt.tech.agent_detected_main_technology` instead
	// True if the agent is running in a IBM CTG process. Not set otherwise.
	// This attribute expects a value of type boolean.
	DtCtgDetected = "dt.ctg.detected"

	// Reports the value of the `DT_CUSTOM_PROP` environment variable as described [here](https://www.dynatrace.com/support/help/how-to-use-dynatrace/process-groups/configuration/define-your-own-process-group-metadata/).
	// This attribute expects a value of type string.
	DtEnvVarsDtCustomProp = "dt.env_vars.dt_custom_prop"

	// Reports the value of the `DT_TAGS` environment variable as described [here](https://www.dynatrace.com/support/help/how-to-use-dynatrace/tags-and-metadata/setup/define-tags-based-on-environment-variables/).
	// This attribute expects a value of type string.
	DtEnvVarsDtTags = "dt.env_vars.dt_tags"

	// Deprecated: Use `dt.tech.agent_detected_main_technology` instead
	// True if the agent is running in an IMS SOAP gateway process. Not set otherwise.
	// This attribute expects a value of type boolean.
	DtImsDetected = "dt.ims.detected"

	// The operating system type.
	// This attribute expects a value of type string from the enumeration DtOsTypeValues.
	DtOsType = "dt.os.type"

	// Human readable (not intended to be parsed) OS version information, like e.g. reported by `ver` or `lsb_release -a` commands.
	// This attribute expects a value of type string.
	DtOsDescription = "dt.os.description"

	// The SNA ID is the Systems Network Architecture (SNA) identifier. It's a unique ID for a given IP address on the network. In combination with the calculated OSI, this is required to calculate the PGI.
	// This attribute expects a value of type string.
	DtHostSnaid = "dt.host.snaid"

	// The SMF ID is the name of the LPAR (logical partition), which is a "host" as far as Dynatrace is concerned. In combination with the IP address, this is required to calculate the OSI.
	// This attribute expects a value of type string.
	DtHostSmfid = "dt.host.smfid"

	// Similar to net.host.ip for spans. Currently, there is no OpenTelemetry convention for IP addresses on resources.
	// This attribute expects a value of type string.
	DtHostIp = "dt.host.ip"

	// The path to the executable.
	// This attribute expects a value of type string.
	DtProcessExecutable = "dt.process.executable"

	// The command line arguments of the process as a string.
	// This attribute expects a value of type string.
	//
	// Ideally, this is the original "raw" command line including the executable path, but this might not be possible in all frameworks. In such cases, this can be a best-effort recreation of the commandline. Examples are truncated for readability.
	DtProcessCommandline = "dt.process.commandline"

	// Process ID (PID) the data belongs to.
	// This attribute expects a value of type string.
	DtProcessPid = "dt.process.pid"

	// The z/OS job name of the process.
	// This attribute expects a value of type string.
	DtProcessZosJobName = "dt.process.zos_job_name"

	// The main technology the agent has detected.
	// This attribute expects a value of type string from the enumeration DtTechAgentDetectedMainTechnologyValues.
	DtTechAgentDetectedMainTechnology = "dt.tech.agent_detected_main_technology"

	// The exporter name. MUST be `odin` for ODIN protocol.
	// This attribute expects a value of type string from the enumeration TelemetryExporterNameValues.
	TelemetryExporterName = "telemetry.exporter.name"

	// The full agent/exporter version.
	// This attribute expects a value of type string.
	TelemetryExporterVersion = "telemetry.exporter.version"

	// The version as exposed to the package manager-.
	// This attribute expects a value of type string.
	//
	// Many package managers won't accept the sprint + timestamp version format as required for `version` if `name` is `odin` but instead need a semver-compatible string with the intended meaning (for example, the `-` used in the sprint version indicates a pre-release version). That package version MAY be provided in this informational attribute.
	TelemetryExporterPackageVersion = "telemetry.exporter.package_version"

	// Name of the WebSphere server.
	// This attribute expects a value of type string.
	DtWebsphereServerName = "dt.websphere.server_name"

	// Name of the WebSphere node the application is running on.
	// This attribute expects a value of type string.
	DtWebsphereNodeName = "dt.websphere.node_name"

	// Name of the WebSphere cell the application is running in.
	// This attribute expects a value of type string.
	DtWebsphereCellName = "dt.websphere.cell_name"

	// Name of the WebSphere cluster the application is running in.
	// This attribute expects a value of type string.
	DtWebsphereClusterName = "dt.websphere.cluster_name"

	// Deprecated: Use `dt.tech.agent_detected_main_technology` instead
	// Set to true if the agent is running in a z/OS Connect EE server process.
	// This attribute expects a value of type boolean.
	//
	// This value is normally either unset or `true`.
	DtZosconnectDetected = "dt.zosconnect.detected"

	// The name of the single function that this runtime instance executes.
	// This attribute expects a value of type string.
	//
	// This is the name of the function as configured/deployed on the FaaS platform and is usually different from the name of the callback function (which may be stored in the [`code.namespace`/`code.function`](../../trace/semantic_conventions/span-general.md#source-code-attributes) span attributes).
	FaasName = "faas.name"

	// The unique ID of the single function that this runtime instance executes.
	// This attribute expects a value of type string.
	//
	// Depending on the cloud provider, use:
	//
	// * **AWS Lambda:** The function [ARN](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html).
	// Take care not to use the "invoked ARN" directly but replace any
	// [alias suffix](https://docs.aws.amazon.com/lambda/latest/dg/configuration-aliases.html) with the resolved function version, as the same runtime instance may be invokable with multiple
	// different aliases.
	// * **GCP:** The [URI of the resource](https://cloud.google.com/iam/docs/full-resource-names)
	// * **Azure:** The [Fully Qualified Resource ID](https://docs.microsoft.com/en-us/rest/api/resources/resources/get-by-id).
	//
	// On some providers, it may not be possible to determine the full ID at startup,
	// which is why this field cannot be made required. For example, on AWS the account ID
	// part of the ARN is not available without calling another AWS API
	// which may be deemed too slow for a short-running lambda function.
	// As an alternative, consider setting `faas.id` as a span attribute instead.
	FaasId = "faas.id"

	// The immutable version of the function being executed.
	// This attribute expects a value of type string.
	//
	// Depending on the cloud provider and platform, use:
	//
	// * **AWS Lambda:** The [function version](https://docs.aws.amazon.com/lambda/latest/dg/configuration-versions.html)
	//   (an integer represented as a decimal string).
	// * **Google Cloud Run:** The [revision](https://cloud.google.com/run/docs/managing/revisions)
	//   (i.e., the function name plus the revision suffix).
	// * **Google Cloud Functions:** The value of the
	//   [`K_REVISION` environment variable](https://cloud.google.com/functions/docs/env-var#runtime_environment_variables_set_automatically).
	// * **Azure Functions:** Not applicable. Do not set this attribute.
	FaasVersion = "faas.version"

	// The execution environment ID as a string, that will be potentially reused for other invocations to the same function/function version.
	// This attribute expects a value of type string.
	//
	// * **AWS Lambda:** Use the (full) log stream name.
	FaasInstance = "faas.instance"

	// The amount of memory available to the serverless function in MiB.
	// This attribute expects a value of type int.
	//
	// It's recommended to set this attribute since e.g. too little memory can easily stop a Java AWS Lambda function from working correctly. On AWS Lambda, the environment variable `AWS_LAMBDA_FUNCTION_MEMORY_SIZE` provides this information.
	FaasMaxMemory = "faas.max_memory"

	// Deprecated: Deprecated. Use attribute `host.name` instead
	// Hostname of the host. It contains what the `hostname` command returns on the host machine.
	// This attribute expects a value of type string.
	HostHostname = "host.hostname"

	// Unique host id.
	// This attribute expects a value of type string.
	//
	// For Cloud this must be the instance_id assigned by the cloud provider.
	HostId = "host.id"

	// Name of the host. It may contain what `hostname` returns on Unix systems, the fully qualified, or a name specified by the user.
	// This attribute expects a value of type string.
	HostName = "host.name"

	// Type of host.
	// This attribute expects a value of type string.
	//
	// For Cloud this must be the machine type.
	HostType = "host.type"

	// Name of the VM image or OS install the host was instantiated from.
	// This attribute expects a value of type string.
	HostImageName = "host.image.name"

	// VM image id.
	// This attribute expects a value of type string.
	//
	// For Cloud, this value is from the provider.
	HostImageId = "host.image.id"

	// The version string of the VM image as defined in [Version Attributes](https://github.com/open-telemetry/opentelemetry-specification/tree/master/specification/resource/semantic_conventions#version-attributes).
	// This attribute expects a value of type string.
	HostImageVersion = "host.image.version"

	// The CPU architecture the host system is running on.
	// This attribute expects a value of type string from the enumeration HostArchValues.
	HostArch = "host.arch"

	// Process identifier (PID).
	// This attribute expects a value of type int.
	ProcessPid = "process.pid"

	// The username of the user that owns the process.
	// This attribute expects a value of type string.
	ProcessOwner = "process.owner"

	// The command used to launch the process (i.e. the command name). On Linux based systems, can be set to the zeroth string in `proc/[pid]/cmdline`. On Windows, can be set to the first parameter extracted from `GetCommandLineW`.
	// This attribute expects a value of type string.
	ProcessCommand = "process.command"

	// The full command used to launch the process. The value can be either a list of strings representing the ordered list of arguments, or a single string representing the full command. On Linux based systems, can be set to the list of null-delimited strings extracted from `proc/[pid]/cmdline`. On Windows, can be set to the result of `GetCommandLineW`.
	// This attribute expects a value of type string[].
	//
	// Union types are not yet supported, so instead of `string` a single-element array is used.
	ProcessCommandLine = "process.command_line"

	// The name of the process executable. On Linux based systems, can be set to the `Name` in `proc/[pid]/status`. On Windows, can be set to the base name of `GetProcessImageFileNameW`.
	// This attribute expects a value of type string.
	ProcessExecutableName = "process.executable.name"

	// The full path to the process executable. On Linux based systems, can be set to the target of `proc/[pid]/exe`. On Windows, can be set to the result of `GetProcessImageFileNameW`.
	// This attribute expects a value of type string.
	ProcessExecutablePath = "process.executable.path"

	// The name of the runtime of this process.
	// This attribute expects a value of type string from the enumeration ProcessRuntimeNameValues.
	//
	// SHOULD be set to one of the values listed below, unless more detailed instructions are provided. If none of the listed values apply, a custom value best describing the runtime CAN be used. For compiled native binaries, this SHOULD be the name of the compiler.
	ProcessRuntimeName = "process.runtime.name"

	// The version of the runtime of this process, as returned by the runtime without modification.
	// This attribute expects a value of type string.
	ProcessRuntimeVersion = "process.runtime.version"

	// An additional description about the runtime of the process, for example a specific vendor customization of the runtime environment.
	// This attribute expects a value of type string.
	ProcessRuntimeDescription = "process.runtime.description"

	// Logical name of the service.
	// This attribute expects a value of type string.
	//
	// MUST be the same for all instances of horizontally scaled services.
	ServiceName = "service.name"

	// A namespace for `service.name`.
	// This attribute expects a value of type string.
	//
	// A string value having a meaning that helps to distinguish a group of services, for example the team name that owns a group of services. `service.name` is expected to be unique within the same namespace. The field is optional. If `service.namespace` is not specified in the Resource then `service.name` is expected to be unique for all services that have no explicit namespace defined (so the empty/unspecified namespace is simply one more valid namespace). Zero-length namespace string is assumed equal to unspecified namespace.
	ServiceNamespace = "service.namespace"

	// The string ID of the service instance.
	// This attribute expects a value of type string.
	//
	// MUST be unique for each instance of the same `service.namespace,service.name` pair (in other words `service.namespace,service.name,service.id` triplet MUST be globally unique). The ID helps to distinguish instances of the same service that exist at the same time (e.g. instances of a horizontally scaled service). It is preferable for the ID to be persistent and stay the same for the lifetime of the service instance, however it is acceptable that the ID is ephemeral and changes during important lifetime events for the service (e.g. service restarts). If the service has no inherent unique ID that can be used as the value of this attribute it is recommended to generate a random Version 1 or Version 4 RFC 4122 UUID (services aiming for reproducible UUIDs may also use Version 5, see RFC 4122 for more recommendations).
	ServiceInstanceId = "service.instance.id"

	// The version string of the service API or implementation as defined in [Version Attributes](https://github.com/open-telemetry/opentelemetry-specification/tree/master/specification/resource/semantic_conventions#version-attributes).
	// This attribute expects a value of type string.
	ServiceVersion = "service.version"

	// The name of the telemetry SDK as defined above.
	// This attribute expects a value of type string.
	//
	// The default OpenTelemetry SDK provided by the OpenTelemetry project MUST set `telemetry.sdk.name` to the value opentelemetry. If another SDK, like a fork or a vendor-provided implementation, is used, this SDK MUST set the attribute `telemetry.sdk.name` to the fully-qualified class or module name of this SDK's main entry point or another suitable identifier depending on the language. The identifier `opentelemetry` is reserved and MUST NOT be used in this case. The identifier SHOULD be stable across different versions of an implementation.
	TelemetrySdkName = "telemetry.sdk.name"

	// The language of the telemetry SDK.
	// This attribute expects a value of type string from the enumeration TelemetrySdkLanguageValues.
	TelemetrySdkLanguage = "telemetry.sdk.language"

	// The version string of the service API or implementation as defined in [Version Attributes](https://github.com/open-telemetry/opentelemetry-specification/tree/master/specification/resource/semantic_conventions#version-attributes).
	// This attribute expects a value of type string.
	TelemetrySdkVersion = "telemetry.sdk.version"

	// The full invoked ARN as provided on the `Context` passed to the function (`Lambda-Runtime-Invoked-Function-Arn` header on the `/runtime/invocation/next` applicable).
	// This attribute expects a value of type string.
	//
	// This may be different from `faas.id` if an alias is involved.
	AwsLambdaInvokedArn = "aws.lambda.invoked_arn"

	// The method or function name, or equivalent (usually rightmost part of the code unit's name).
	// This attribute expects a value of type string.
	CodeFunction = "code.function"

	// The "namespace" within which `code.function` is defined. Usually the qualified class or module name, such that `code.namespace` + some separator + `code.function` form a unique identifier for the code unit.
	// This attribute expects a value of type string.
	CodeNamespace = "code.namespace"

	// The source code file name that identifies the code unit as uniquely as possible (preferably an absolute file path).
	// This attribute expects a value of type string.
	CodeFilepath = "code.filepath"

	// The line number in `code.filepath` best representing the operation. It SHOULD point within the code unit named in code.function.
	// This attribute expects a value of type int.
	CodeLineno = "code.lineno"

	// An identifier for the database management system (DBMS) product being used. See below for a list of well-known identifiers.
	// This attribute expects a value of type string from the enumeration DbSystemValues.
	DbSystem = "db.system"

	// The connection string used to connect to the database.
	// This attribute expects a value of type string.
	//
	// It is recommended to remove embedded credentials.
	DbConnectionString = "db.connection_string"

	// Username for accessing the database.
	// This attribute expects a value of type string.
	DbUser = "db.user"

	// The fully-qualified class name of the [Java Database Connectivity (JDBC)](https://docs.oracle.com/javase/8/docs/technotes/guides/jdbc/) driver used to connect.
	// This attribute expects a value of type string.
	DbJdbcDriverClassname = "db.jdbc.driver_classname"

	// If no tech-specific attribute is defined, this attribute is used to report the name of the database being accessed. For commands that switch the database, this should be set to the target database (even if the command fails).
	// This attribute expects a value of type string.
	//
	// In some SQL databases, the database name to be used is called "schema name".
	DbName = "db.name"

	// The database statement being executed.
	// This attribute expects a value of type string.
	//
	// The value may be sanitized to exclude sensitive information.
	DbStatement = "db.statement"

	// The name of the operation being executed, e.g. the [MongoDB command name](https://docs.mongodb.com/manual/reference/command/#database-operations) such as `findAndModify`.
	// This attribute expects a value of type string.
	//
	// While it would semantically make sense to set this, e.g., to a SQL keyword like `SELECT` or `INSERT`, it is not recommended to attempt any client-side parsing of `db.statement` just to get this property (the back end can do that if required).
	DbOperation = "db.operation"

	// Remote hostname or similar, see note below.
	// This attribute expects a value of type string.
	NetPeerName = "net.peer.name"

	// Remote address of the peer (dotted decimal for IPv4 or [RFC5952](https://rfc-editor.org/rfc/rfc5952) for IPv6).
	// This attribute expects a value of type string.
	NetPeerIp = "net.peer.ip"

	// Remote port number.
	// This attribute expects a value of type int.
	NetPeerPort = "net.peer.port"

	// Transport protocol used. See note below.
	// This attribute expects a value of type string from the enumeration NetTransportValues.
	NetTransport = "net.transport"

	// The Microsoft SQL Server [instance name](https://docs.microsoft.com/en-us/sql/connect/jdbc/building-the-connection-url?view=sql-server-ver15) connecting to. This name is used to determine the port of a named instance.
	// This attribute expects a value of type string.
	//
	// If setting a `db.mssql.instance_name`, `net.peer.port` is no longer required (but still recommended if non-standard).
	DbMssqlInstanceName = "db.mssql.instance_name"

	// The name of the keyspace being accessed. To be used instead of the generic `db.name` attribute.
	// This attribute expects a value of type string.
	DbCassandraKeyspace = "db.cassandra.keyspace"

	// The [HBase namespace](https://hbase.apache.org/book.html#_namespace) being accessed. To be used instead of the generic `db.name` attribute.
	// This attribute expects a value of type string.
	DbHbaseNamespace = "db.hbase.namespace"

	// The index of the database being accessed as used in the [`SELECT` command](https://redis.io/commands/select), provided as an integer. To be used instead of the generic `db.name` attribute.
	// This attribute expects a value of type int.
	DbRedisDatabaseIndex = "db.redis.database_index"

	// The collection being accessed within the database stated in `db.name`.
	// This attribute expects a value of type string.
	DbMongodbCollection = "db.mongodb.collection"

	// URL of the gateway.
	// This attribute expects a value of type string.
	DtCtgGatewayurl = "dt.ctg.gatewayurl"

	// Type of the CTG GatewayRequest.
	// This attribute expects a value of type string from the enumeration DtCtgRequesttypeValues.
	DtCtgRequesttype = "dt.ctg.requesttype"

	// Integer representing the specific calltype of the CTG GatewayRequest.
	// This attribute expects a value of type int from the enumeration DtCtgCalltypeValues.
	DtCtgCalltype = "dt.ctg.calltype"

	// ID/name of the server.
	// This attribute expects a value of type string.
	DtCtgServerid = "dt.ctg.serverid"

	// ID/name of the user.
	// This attribute expects a value of type string.
	DtCtgUserid = "dt.ctg.userid"

	// ID of the transaction.
	// This attribute expects a value of type string.
	DtCtgTransid = "dt.ctg.transid"

	// Name of the CICS program.
	// This attribute expects a value of type string.
	DtCtgProgram = "dt.ctg.program"

	// Length of the communication area.
	// This attribute expects a value of type int.
	DtCtgCommarealength = "dt.ctg.commarealength"

	// See "ExtendModes" section below.
	// This attribute expects a value of type int.
	DtCtgExtendmode = "dt.ctg.extendmode"

	// Name of the terminal resource.
	// This attribute expects a value of type string.
	DtCtgTermid = "dt.ctg.termid"

	// CTG response code.
	// This attribute expects a value of type int.
	DtCtgRc = "dt.ctg.rc"

	// The topology of the database in relation to the application performing database requests.
	// This attribute expects a value of type string from the enumeration DtDbTopologyValues.
	DtDbTopology = "dt.db.topology"

	// The exception types of a caused-by chain, represented by their fully-qualified type names encoded as a single string (see [Encoding of Exception Data](https://bitbucket.lab.dynatrace.org/projects/ODIN/repos/odin-spec/browse/spec/semantic_conventions/exception_conventions.md#encoding-of-exception-data)).
	// This attribute expects a value of type string.
	DtExceptionTypes = "dt.exception.types"

	// Messages providing details about the exceptions of a caused-by chain encoded as a single string (see [Encoding of Exception Data](https://bitbucket.lab.dynatrace.org/projects/ODIN/repos/odin-spec/browse/spec/semantic_conventions/exception_conventions.md#encoding-of-exception-data)).
	// This attribute expects a value of type string.
	DtExceptionMessages = "dt.exception.messages"

	// Stack traces for all exceptions in a caused-by chain, serialized into a single string (see [Encoding of Exception Data](https://bitbucket.lab.dynatrace.org/projects/ODIN/repos/odin-spec/browse/spec/semantic_conventions/exception_conventions.md#encoding-of-exception-data)).
	// This attribute expects a value of type string.
	DtExceptionSerializedStacktraces = "dt.exception.serialized_stacktraces"

	// Type of the trigger on which the function is executed.
	// This attribute expects a value of type string from the enumeration FaasTriggerValues.
	FaasTrigger = "faas.trigger"

	// The execution ID of the current function execution.
	// This attribute expects a value of type string.
	FaasExecution = "faas.execution"

	// The `X-Amzn-Trace-Id` HTTP response header for [AWS X-Ray](https://docs.aws.amazon.com/xray/latest/devguide/aws-xray.html) tracing.
	// This attribute expects a value of type string.
	DtFaasAwsXAmznTraceId = "dt.faas.aws.x_amzn_trace_id"

	// The AWS `X-Amzn-RequestId` HTTP response header.
	// This attribute expects a value of type string.
	//
	// The different notation of the `RequestId` part (missing `-` character) compared to `X-Amzn-Trace-Id` is intentional.
	DtFaasAwsXAmznRequestId = "dt.faas.aws.x_amzn_request_id"

	// Deprecated: Use `code.function` instead
	// The method or function name, or equivalent (usually rightmost part of the code unit's name).
	// This attribute expects a value of type string.
	DtCodeFunc = "dt.code.func"

	// Deprecated: Use `code.namespace` instead
	// The "namespace" within which `dt.code.func` is defined. Usually the qualified class or module name, such that `dt.code.ns` + some separator + `dt.code.func` form a unique identifier for the code unit.
	// This attribute expects a value of type string.
	DtCodeNs = "dt.code.ns"

	// Deprecated: Use `code.filepath` instead
	// The source code file name that identifies the code unit as uniquely as possible (preferably an absolute file path).
	// This attribute expects a value of type string.
	DtCodeFilepath = "dt.code.filepath"

	// Deprecated: Use `code.lineno` instead
	// The line number in `dt.code.filepath` best representing the operation. It SHOULD point within the code unit named in `dt.code.func`.
	// This attribute expects a value of type int.
	DtCodeLineno = "dt.code.lineno"

	// Full stacktrace of call to Span.Start (possibly including call to span processor's OnStart).
	// This attribute expects a value of type string.
	DtStacktraceOnstart = "dt.stacktrace.onstart"

	// Full stacktrace of call to Span.End (possibly including call to span processor's OnEnd).
	// This attribute expects a value of type string.
	DtStacktraceOnend = "dt.stacktrace.onend"

	// HTTP request method.
	// This attribute expects a value of type string.
	HttpMethod = "http.method"

	// Full HTTP request URL in the form `scheme://host[:port]/path?query[#fragment]`. Usually the fragment is not transmitted over HTTP, but if it is known, it should be included nevertheless.
	// This attribute expects a value of type string.
	HttpUrl = "http.url"

	// The full request target as passed in a HTTP request line or equivalent.
	// This attribute expects a value of type string.
	HttpTarget = "http.target"

	// The value of the [HTTP host header](https://rfc-editor.org/rfc/rfc7230#section-5.4). When the header is empty or not present, this attribute should be the same.
	// This attribute expects a value of type string.
	HttpHost = "http.host"

	// The URI scheme identifying the used protocol.
	// This attribute expects a value of type string.
	HttpScheme = "http.scheme"

	// [HTTP response status code](https://rfc-editor.org/rfc/rfc7231#section-6).
	// This attribute expects a value of type int.
	HttpStatusCode = "http.status_code"

	// [HTTP reason phrase](https://rfc-editor.org/rfc/rfc7230#section-3.1.2).
	// This attribute expects a value of type string.
	HttpStatusText = "http.status_text"

	// Kind of HTTP protocol used.
	// This attribute expects a value of type string from the enumeration HttpFlavorValues.
	//
	// If `net.transport` is not specified, it can be assumed to be `IP.TCP` except if `http.flavor` is `QUIC`, in which case `IP.UDP` is assumed.
	HttpFlavor = "http.flavor"

	// Value of the [HTTP User-Agent](https://rfc-editor.org/rfc/rfc7231#section-5.5.3) header sent by the client.
	// This attribute expects a value of type string.
	HttpUserAgent = "http.user_agent"

	// The size of the request payload body in bytes. This is the number of bytes transferred excluding headers and is often, but not always, present as the [Content-Length](https://rfc-editor.org/rfc/rfc7230#section-3.3.2) header. For requests using transport encoding, this should be the compressed size.
	// This attribute expects a value of type int.
	HttpRequestContentLength = "http.request_content_length"

	// The size of the uncompressed request payload body after transport decoding. Not set if transport encoding not used.
	// This attribute expects a value of type int.
	HttpRequestContentLengthUncompressed = "http.request_content_length_uncompressed"

	// The size of the response payload body in bytes. This is the number of bytes transferred excluding headers and is often, but not always, present as the [Content-Length](https://rfc-editor.org/rfc/rfc7230#section-3.3.2) header. For requests using transport encoding, this should be the compressed size.
	// This attribute expects a value of type int.
	HttpResponseContentLength = "http.response_content_length"

	// The size of the uncompressed response payload body after transport decoding. Not set if transport encoding not used.
	// This attribute expects a value of type int.
	HttpResponseContentLengthUncompressed = "http.response_content_length_uncompressed"

	// Like `net.peer.ip` but for the host IP. Useful in case of a multi-IP host.
	// This attribute expects a value of type string.
	NetHostIp = "net.host.ip"

	// Like `net.peer.port` but for the host port.
	// This attribute expects a value of type int.
	NetHostPort = "net.host.port"

	// Local hostname or similar, see note below.
	// This attribute expects a value of type string.
	NetHostName = "net.host.name"

	// The primary server name of the matched virtual host. This should be obtained via configuration. If no such configuration can be obtained, this attribute MUST NOT be set ( `net.host.name` should be used instead).
	// This attribute expects a value of type string.
	//
	// http.url is usually not readily available on the server side but would have to be assembled in a cumbersome and sometimes lossy process from other information (see e.g. open-telemetry/opentelemetry-python/pull/148). It is thus preferred to supply the raw data that is available.
	HttpServerName = "http.server_name"

	// The matched route (path template).
	// This attribute expects a value of type string.
	HttpRoute = "http.route"

	// The IP address of the original client behind all proxies, if known (e.g. from [X-Forwarded-For](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)).
	// This attribute expects a value of type string.
	//
	// This is not necessarily the same as `net.peer.ip`, which would identify the network-level peer, which may be a proxy.
	HttpClientIp = "http.client_ip"

	// Tech-dependent, eg. servlet context.
	// This attribute expects a value of type string.
	DtHttpApplicationId = "dt.http.application_id"

	// The path prefix of the URL that identifies this HTTP application. If multiple roots exist, the one that was matched for this request should be used.
	// This attribute expects a value of type string.
	//
	// See [OTel definition](https://github.com/open-telemetry/opentelemetry-specification/blob/v0.6.0/specification/trace/semantic_conventions/http.md#http-server-definitions).
	DtHttpContextRoot = "dt.http.context_root"

	// `Referer` header.
	// This attribute expects a value of type string.
	DtHttpRequestHeaderReferer = "dt.http.request.header.referer"

	// `X-Dynatrace-Test` header.
	// This attribute expects a value of type string.
	DtHttpRequestHeaderXDynatraceTest = "dt.http.request.header.x-dynatrace-test"

	// `X-Dynatrace-Tenant` header.
	// This attribute expects a value of type string.
	DtHttpRequestHeaderXDynatraceTenant = "dt.http.request.header.x-dynatrace-tenant"

	// `Forwarded` header.
	// This attribute expects a value of type string.
	DtHttpRequestHeaderForwarded = "dt.http.request.header.forwarded"

	// `X-Forwarded-For` header (only required if `dt.http.request.header.forwarded` isn't provided).
	// This attribute expects a value of type string.
	DtHttpRequestHeaderXForwardedFor = "dt.http.request.header.x-forwarded-for"

	// Set to true for an interaction with IMS SOAP Gateway.
	// This attribute expects a value of type boolean.
	//
	// This value is normally either unset or `true`.
	DtImsIsIms = "dt.ims.is_ims"

	// Contains the instrumentation library name.
	// This attribute expects a value of type string.
	OtelLibraryName = "otel.library.name"

	// Contains the instrumentation library version.
	// This attribute expects a value of type string.
	OtelLibraryVersion = "otel.library.version"

	// A string identifying the messaging system.
	// This attribute expects a value of type string.
	MessagingSystem = "messaging.system"

	// The message destination name. This might be equal to the span name but is required nevertheless.
	// This attribute expects a value of type string.
	MessagingDestination = "messaging.destination"

	// The kind of message destination.
	// This attribute expects a value of type string from the enumeration MessagingDestinationKindValues.
	MessagingDestinationKind = "messaging.destination_kind"

	// A boolean that is true if the message destination is temporary.
	// This attribute expects a value of type boolean.
	MessagingTempDestination = "messaging.temp_destination"

	// The name of the transport protocol.
	// This attribute expects a value of type string.
	MessagingProtocol = "messaging.protocol"

	// The version of the transport protocol.
	// This attribute expects a value of type string.
	MessagingProtocolVersion = "messaging.protocol_version"

	// Connection string.
	// This attribute expects a value of type string.
	MessagingUrl = "messaging.url"

	// A value used by the messaging system as an identifier for the message, represented as a string.
	// This attribute expects a value of type string.
	MessagingMessageId = "messaging.message_id"

	// A value identifying the conversation to which the message belongs, represented as a string. Sometimes called "Correlation ID".
	// This attribute expects a value of type string.
	MessagingConversationId = "messaging.conversation_id"

	// The (uncompressed) size of the message payload in bytes. Also use this attribute if it is unknown whether the compressed or uncompressed payload size is reported.
	// This attribute expects a value of type int.
	MessagingMessagePayloadSizeBytes = "messaging.message_payload_size_bytes"

	// The compressed size of the message payload in bytes.
	// This attribute expects a value of type int.
	MessagingMessagePayloadCompressedSizeBytes = "messaging.message_payload_compressed_size_bytes"

	// Name of IBM's queue manager.
	// This attribute expects a value of type string.
	//
	// Only for IBM MQ spans.
	DtMessagingIbmQueuemanagerName = "dt.messaging.ibm.queuemanager.name"

	// The type of content of the jms message.
	// This attribute expects a value of type string from the enumeration DtMessagingJmsMessageTypeValues.
	DtMessagingJmsMessageType = "dt.messaging.jms.message_type"

	// Number of messages being sent/received/processed at once.
	// This attribute expects a value of type int.
	//
	// This batch size attribute is kind of a workaround for situations where we cannot better distinguish the individual messages.
	DtMessagingBatchSize = "dt.messaging.batch_size"

	// MUST be set to `true` if this is a link that would have been the parent but was suppressed by the OpenTelemetry integration.
	// This attribute expects a value of type boolean.
	DtParentIsSuppressedPrimary = "dt.parent.is_suppressed_primary"

	// Value of x-dtc header.
	// This attribute expects a value of type string.
	DtRumDtc = "dt.rum.dtc"

	// Monitored entity id of configured application.
	// This attribute expects a value of type string.
	DtRumAppMeId = "dt.rum.app_me_id"

	// Name of the header in `dt.http.request.headers` to parse clientip from.
	// This attribute expects a value of type string.
	DtRumClientipHeaderName = "dt.rum.clientip_header_name"

	// The name of the z/OS application program called by the request.
	// This attribute expects a value of type string.
	DtZosconnectProgram = "dt.zosconnect.program"

	// The z/OS Connect request ID.
	// This attribute expects a value of type int.
	DtZosconnectRequestId = "dt.zosconnect.request_id"

	// The service provider name.
	// This attribute expects a value of type string.
	DtZosconnectServiceProviderName = "dt.zosconnect.service_provider_name"

	// The system of record reference.
	// This attribute expects a value of type string.
	DtZosconnectSorReference = "dt.zosconnect.sor_reference"

	// The system of record identifier.  The format differs depending on the SOR type.
	// This attribute expects a value of type string.
	//
	// <https://www.ibm.com/support/knowledgecenter/SS4SVW_3.0.0/javadoc/com/ibm/zosconnect/spi/Data.html?view=embed#SOR_IDENTIFIER>.
	DtZosconnectSorIdentifier = "dt.zosconnect.sor_identifier"

	// Identifier for the resource invoked on the system of record. The format differs depending on the SOR type.
	// This attribute expects a value of type string.
	//
	// <https://www.ibm.com/support/knowledgecenter/SS4SVW_3.0.0/javadoc/com/ibm/zosconnect/spi/Data.html?view=embed#SOR_RESOURCE>.
	DtZosconnectSorResource = "dt.zosconnect.sor_resource"

	// The system of record type.
	// This attribute expects a value of type string from the enumeration DtZosconnectSorTypeValues.
	DtZosconnectSorType = "dt.zosconnect.sor_type"

	// The length of the request payload in bytes.
	// This attribute expects a value of type int.
	DtZosconnectInputPayloadLength = "dt.zosconnect.input_payload_length"

	// The length of the response payload in bytes.
	// This attribute expects a value of type int.
	DtZosconnectOutputPayloadLength = "dt.zosconnect.output_payload_length"

	// The z/OS Connect API name.
	// This attribute expects a value of type string.
	DtZosconnectApiName = "dt.zosconnect.api_name"

	// The z/OS Connect service name.
	// This attribute expects a value of type string.
	DtZosconnectServiceName = "dt.zosconnect.service_name"

	// The z/OS Connect API description.
	// This attribute expects a value of type string.
	DtZosconnectApiDescription = "dt.zosconnect.api_description"

	// The z/OS Connect service description.
	// This attribute expects a value of type string.
	DtZosconnectServiceDescription = "dt.zosconnect.service_description"

	// The z/OS Connect API version.
	// This attribute expects a value of type string.
	DtZosconnectApiVersion = "dt.zosconnect.api_version"

	// The z/OS Connect service version.
	// This attribute expects a value of type string.
	DtZosconnectServiceVersion = "dt.zosconnect.service_version"

	// The type of the REST request.
	// This attribute expects a value of type string from the enumeration DtZosconnectRequestTypeValues.
	//
	// <https://www.ibm.com/support/knowledgecenter/SS4SVW_3.0.0/javadoc/com/ibm/zosconnect/spi/Data.RequestType.html>.
	DtZosconnectRequestType = "dt.zosconnect.request_type"

	// The name of the source on which the triggering operation was performed. For example, in Cloud Storage or S3 corresponds to the bucket name, and in Cosmos DB to the database name.
	// This attribute expects a value of type string.
	FaasDocumentCollection = "faas.document.collection"

	// Describes the type of the operation that was performed on the data.
	// This attribute expects a value of type string from the enumeration FaasDocumentOperationValues.
	FaasDocumentOperation = "faas.document.operation"

	// A string containing the time when the data was accessed in the [ISO 8601](https://www.iso.org/iso-8601-date-and-time-format.html) format expressed in [UTC](https://www.w3.org/TR/NOTE-datetime).
	// This attribute expects a value of type string.
	FaasDocumentTime = "faas.document.time"

	// The document name/table subjected to the operation. For example, in Cloud Storage or S3 is the name of the file, and in Cosmos DB the table name.
	// This attribute expects a value of type string.
	FaasDocumentName = "faas.document.name"

	// A string containing the function invocation time in the [ISO 8601](https://www.iso.org/iso-8601-date-and-time-format.html) format expressed in [UTC](https://www.w3.org/TR/NOTE-datetime).
	// This attribute expects a value of type string.
	FaasTime = "faas.time"

	// A string containing the schedule period as [Cron Expression](https://docs.oracle.com/cd/E12058_01/doc/doc.1014/e12030/cron_expressions.htm).
	// This attribute expects a value of type string.
	FaasCron = "faas.cron"

	// A boolean that is true if the serverless function is executed for the first time (aka cold-start).
	// This attribute expects a value of type boolean.
	FaasColdstart = "faas.coldstart"

	// The name of the invoked function.
	// This attribute expects a value of type string.
	//
	// SHOULD be equal to the `faas.name` resource attribute of the invoked function.
	FaasInvokedName = "faas.invoked_name"

	// The cloud provider of the invoked function.
	// This attribute expects a value of type string from the enumeration FaasInvokedProviderValues.
	//
	// SHOULD be equal to the `cloud.provider` resource attribute of the invoked function.
	FaasInvokedProvider = "faas.invoked_provider"

	// The cloud region of the invoked function.
	// This attribute expects a value of type string.
	//
	// SHOULD be equal to the `cloud.region` resource attribute of the invoked function.
	FaasInvokedRegion = "faas.invoked_region"

	// Username or client_id extracted from the access token or Authorization header in the inbound request from outside the system.
	// This attribute expects a value of type string.
	EnduserId = "enduser.id"

	// Actual/assumed role the client is making the request under extracted from token or application security context.
	// This attribute expects a value of type string.
	EnduserRole = "enduser.role"

	// Scopes or granted authorities the client currently possesses extracted from token or application security context. The value would come from the scope associated with an OAuth 2.0 Access Token or an attribute value in a SAML 2.0 Assertion.
	// This attribute expects a value of type string.
	EnduserScope = "enduser.scope"

	// The value `aws-api`.
	// This attribute expects a value of type string from the enumeration RpcSystemValues.
	RpcSystem = "rpc.system"

	// The name of the service to which a request is made, as returned by the AWS SDK.
	// This attribute expects a value of type string.
	//
	// This is the logical name of the service from the RPC interface perspective, which can be different from the name of any implementing class. The `code.namespace` attribute may be used to store the latter (despite the attribute name, it may include a class name; e.g., class with method actually executing the call on the server side, RPC client stub class on the client side).
	RpcService = "rpc.service"

	// The name of the operation corresponding to the request, as returned by the AWS SDK.
	// This attribute expects a value of type string.
	//
	// This is the logical name of the method from the RPC interface perspective, which can be different from the name of any implementing method/function. The `code.function` attribute may be used to store the latter (e.g., method actually executing the call on the server side, RPC client stub method on the client side).
	RpcMethod = "rpc.method"

	// The keys in the `RequestItems` object field.
	// This attribute expects a value of type string[].
	AwsDynamodbTableNames = "aws.dynamodb.table_names"

	// The JSON-serialized value of each item in the `ConsumedCapacity` response field.
	// This attribute expects a value of type string[].
	AwsDynamodbConsumedCapacity = "aws.dynamodb.consumed_capacity"

	// The JSON-serialized value of the `ItemCollectionMetrics` response field.
	// This attribute expects a value of type string.
	AwsDynamodbItemCollectionMetrics = "aws.dynamodb.item_collection_metrics"

	// The value of the `ProvisionedThroughput.ReadCapacityUnits` request parameter.
	// This attribute expects a value of type double.
	AwsDynamodbProvisionedReadCapacity = "aws.dynamodb.provisioned_read_capacity"

	// The value of the `ProvisionedThroughput.WriteCapacityUnits` request parameter.
	// This attribute expects a value of type double.
	AwsDynamodbProvisionedWriteCapacity = "aws.dynamodb.provisioned_write_capacity"

	// The value of the `ConsistentRead` request parameter.
	// This attribute expects a value of type boolean.
	AwsDynamodbConsistentRead = "aws.dynamodb.consistent_read"

	// The value of the `ProjectionExpression` request parameter.
	// This attribute expects a value of type string.
	AwsDynamodbProjection = "aws.dynamodb.projection"

	// The value of the `Limit` request parameter.
	// This attribute expects a value of type int.
	AwsDynamodbLimit = "aws.dynamodb.limit"

	// The value of the `AttributesToGet` request parameter.
	// This attribute expects a value of type string[].
	AwsDynamodbAttributesToGet = "aws.dynamodb.attributes_to_get"

	// The value of the `IndexName` request parameter.
	// This attribute expects a value of type string.
	AwsDynamodbIndexName = "aws.dynamodb.index_name"

	// The value of the `Select` request parameter.
	// This attribute expects a value of type string.
	AwsDynamodbSelect = "aws.dynamodb.select"

	// The JSON-serialized value of each item of the `GlobalSecondaryIndexes` request field.
	// This attribute expects a value of type string[].
	AwsDynamodbGlobalSecondaryIndexes = "aws.dynamodb.global_secondary_indexes"

	// The JSON-serialized value of each item of the `LocalSecondaryIndexes` request field.
	// This attribute expects a value of type string[].
	AwsDynamodbLocalSecondaryIndexes = "aws.dynamodb.local_secondary_indexes"

	// The value of the `ExclusiveStartTableName` request parameter.
	// This attribute expects a value of type string.
	AwsDynamodbExclusiveStartTable = "aws.dynamodb.exclusive_start_table"

	// The the number of items in the `TableNames` response parameter.
	// This attribute expects a value of type int.
	AwsDynamodbTableCount = "aws.dynamodb.table_count"

	// The value of the `ScanIndexForward` request parameter.
	// This attribute expects a value of type boolean.
	AwsDynamodbScanForward = "aws.dynamodb.scan_forward"

	// The value of the `Segment` request parameter.
	// This attribute expects a value of type int.
	AwsDynamodbSegment = "aws.dynamodb.segment"

	// The value of the `TotalSegments` request parameter.
	// This attribute expects a value of type int.
	AwsDynamodbTotalSegments = "aws.dynamodb.total_segments"

	// The value of the `Count` response parameter.
	// This attribute expects a value of type int.
	AwsDynamodbCount = "aws.dynamodb.count"

	// The value of the `ScannedCount` response parameter.
	// This attribute expects a value of type int.
	AwsDynamodbScannedCount = "aws.dynamodb.scanned_count"

	// The JSON-serialized value of each item in the `AttributeDefinitions` request field.
	// This attribute expects a value of type string[].
	AwsDynamodbAttributeDefinitions = "aws.dynamodb.attribute_definitions"

	// The JSON-serialized value of each item in the the `GlobalSecondaryIndexUpdates` request field.
	// This attribute expects a value of type string[].
	AwsDynamodbGlobalSecondaryIndexUpdates = "aws.dynamodb.global_secondary_index_updates"

	// The cloud region of the invoked resource.
	// This attribute expects a value of type string.
	//
	// SHOULD be equal to the `cloud.region` resource attribute of the invoked resource.
	AwsRegion = "aws.region"

	// A string identifying which part and kind of message consumption this span describes.
	// This attribute expects a value of type string from the enumeration MessagingOperationValues.
	MessagingOperation = "messaging.operation"

	// The [numeric status code](https://github.com/grpc/grpc/blob/v1.33.2/doc/statuscodes.md) of the gRPC request.
	// This attribute expects a value of type int from the enumeration RpcGrpcStatusCodeValues.
	RpcGrpcStatusCode = "rpc.grpc.status_code"

	// Protocol version as in `jsonrpc` property of request/response. Since JSON-RPC 1.0 does not specify this, the value can be omitted.
	// This attribute expects a value of type string.
	RpcJsonrpcVersion = "rpc.jsonrpc.version"

	// `id` property of request or response. Since protocol allows id to be int, string, `null` or missing (for notifications), value is expected to be cast to string for simplicity. Use empty string in case of `null` value. Omit entirely if this is a notification.
	// This attribute expects a value of type string.
	RpcJsonrpcRequestId = "rpc.jsonrpc.request_id"

	// `error.code` property of response if it is an error response.
	// This attribute expects a value of type int.
	RpcJsonrpcErrorCode = "rpc.jsonrpc.error_code"

	// `error.message` property of response if it is an error response.
	// This attribute expects a value of type string.
	RpcJsonrpcErrorMessage = "rpc.jsonrpc.error_message"

	// Whether this is a received or sent message.
	// This attribute expects a value of type string from the enumeration MessageTypeValues.
	MessageType = "message.type"

	// MUST be calculated as two different counters starting from `1` one for sent messages and one for received message.
	// This attribute expects a value of type int.
	//
	// This way we guarantee that the values will be consistent between different implementations.
	MessageId = "message.id"

	// Compressed size of the message in bytes.
	// This attribute expects a value of type int.
	MessageCompressedSize = "message.compressed_size"

	// Uncompressed size of the message in bytes.
	// This attribute expects a value of type int.
	MessageUncompressedSize = "message.uncompressed_size"
)

// Enum definitions

// The available values for cloud.provider.
const (
	// Amazon Web Services
	CloudProviderAws = "aws"

	// Microsoft Azure
	CloudProviderAzure = "azure"

	// Google Cloud Platform
	CloudProviderGcp = "gcp"
)

// The available values for cloud.platform.
const (
	// AWS Elastic Compute Cloud
	CloudPlatformAwsEc2 = "aws_ec2"

	// AWS Elastic Container Service
	CloudPlatformAwsEcs = "aws_ecs"

	// AWS Elastic Kubernetes Service
	CloudPlatformAwsEks = "aws_eks"

	// AWS Lambda
	CloudPlatformAwsLambda = "aws_lambda"

	// AWS Elastic Beanstalk
	CloudPlatformAwsElasticBeanstalk = "aws_elastic_beanstalk"

	// Azure Virtual Machines
	CloudPlatformAzureVm = "azure_vm"

	// Azure Container Instances
	CloudPlatformAzureContainerInstances = "azure_container_instances"

	// Azure Kubernetes Service
	CloudPlatformAzureAks = "azure_aks"

	// Azure Functions
	CloudPlatformAzureFunctions = "azure_functions"

	// Azure App Service
	CloudPlatformAzureAppService = "azure_app_service"

	// Google Cloud Compute Engine (GCE)
	CloudPlatformGcpComputeEngine = "gcp_compute_engine"

	// Google Cloud Run
	CloudPlatformGcpCloudRun = "gcp_cloud_run"

	// Google Cloud Kubernetes Engine (GKE)
	CloudPlatformGcpKubernetesEngine = "gcp_kubernetes_engine"

	// Google Cloud Functions (GCF)
	CloudPlatformGcpCloudFunctions = "gcp_cloud_functions"

	// Google Cloud App Engine (GAE)
	CloudPlatformGcpAppEngine = "gcp_app_engine"
)

// The available values for dt.os.type.
const (
	// Fallback, if the OS can't be determined
	DtOsTypeUnknown = "UNKNOWN"

	// Microsoft Windows
	DtOsTypeWindows = "WINDOWS"

	// Linux
	DtOsTypeLinux = "LINUX"

	// HP-UX (Hewlett Packard Unix)
	DtOsTypeHpux = "HPUX"

	// AIX (Advanced Interactive eXecutive)
	DtOsTypeAix = "AIX"

	// Oracle Solaris
	DtOsTypeSolaris = "SOLARIS"

	// IBM z/OS
	DtOsTypeZos = "ZOS"

	// Apple Darwin
	DtOsTypeDarwin = "DARWIN"
)

// The available values for dt.tech.agent_detected_main_technology.
const (
	// Fallback, if the main technology can't be determined
	DtTechAgentDetectedMainTechnologyUnknown = "unknown"

	// Agent is monitoring an AWS Lambda function (set if and only if `faas.name` attribute is present)
	DtTechAgentDetectedMainTechnologyAwsLambda = "aws_lambda"

	// Agent is monitoring a Z/OS Connect instance
	DtTechAgentDetectedMainTechnologyZosConnect = "zos_connect"

	// Agent is monitoring a CICS Transaction Gateway instance
	DtTechAgentDetectedMainTechnologyCtg = "ctg"

	// Agent is monitoring a IMS SOAP Gateway instance
	DtTechAgentDetectedMainTechnologyIms = "ims"

	// Agent is monitoring a WebSphere Application Server instance that is not monitored as a more specific technology (set only if any of `dt.websphere.node_name`, `dt.websphere.node_name`, `dt.websphere.cell_name` or `dt.websphere.cluster_name` is present, if no more specific technology, like `zos_connect` applies)
	DtTechAgentDetectedMainTechnologyWebsphereAs = "websphere_as"

	// Agent is monitoring a WebSphere Liberty instance that is not monitored as a more specific technology (set only if `dt.websphere.server` is present, if no more specific technology, like `zos_connect` applies)
	DtTechAgentDetectedMainTechnologyWebsphereLiberty = "websphere_liberty"
)

// The available values for telemetry.exporter.name.
const (
	// ODIN exporter. If using this, `version` MUST be in the format described in [Versioning](../../versioning.md)
	TelemetryExporterNameOdin = "odin"
)

// The available values for host.arch.
const (
	// AMD64
	HostArchAmd64 = "amd64"

	// ARM32
	HostArchArm32 = "arm32"

	// ARM64
	HostArchArm64 = "arm64"

	// Itanium
	HostArchIa64 = "ia64"

	// 32-bit PowerPC
	HostArchPpc32 = "ppc32"

	// 64-bit PowerPC
	HostArchPpc64 = "ppc64"

	// IBM z/Architecture
	HostArchS390x = "s390x"

	// 32-bit x86
	HostArchX86 = "x86"
)

// The available values for process.runtime.name.
const (
	// beam
	ProcessRuntimeNameBeam = "BEAM"

	// gc
	ProcessRuntimeNameGc = "Go compiler"

	// gccgo
	ProcessRuntimeNameGccgo = "GCC Go frontend"

	// nodejs
	ProcessRuntimeNameNodejs = "NodeJS"

	// browser
	ProcessRuntimeNameBrowser = "Web Browser"

	// iojs
	ProcessRuntimeNameIojs = "io.js"

	// dotnet-core
	ProcessRuntimeNameDotnetCore = ".NET Core, .NET 5+"

	// dotnet-framework
	ProcessRuntimeNameDotnetFramework = ".NET Framework"

	// mono
	ProcessRuntimeNameMono = "Mono"

	// cpython
	ProcessRuntimeNameCpython = "CPython"

	// ironpython
	ProcessRuntimeNameIronpython = "IronPython"

	// jython
	ProcessRuntimeNameJython = "Jython"

	// pypy
	ProcessRuntimeNamePypy = "PyPy"

	// pythonnet
	ProcessRuntimeNamePythonnet = "PythonNet"
)

// The available values for telemetry.sdk.language.
const (
	// cpp
	TelemetrySdkLanguageCpp = "cpp"

	// dotnet
	TelemetrySdkLanguageDotnet = "dotnet"

	// erlang
	TelemetrySdkLanguageErlang = "erlang"

	// go
	TelemetrySdkLanguageGo = "go"

	// java
	TelemetrySdkLanguageJava = "java"

	// nodejs
	TelemetrySdkLanguageNodejs = "nodejs"

	// php
	TelemetrySdkLanguagePhp = "php"

	// python
	TelemetrySdkLanguagePython = "python"

	// ruby
	TelemetrySdkLanguageRuby = "ruby"

	// webjs
	TelemetrySdkLanguageWebjs = "webjs"
)

// The available values for db.system.
const (
	// Some other SQL database. Fallback only. See notes
	DbSystemOtherSql = "other_sql"

	// Microsoft SQL Server
	DbSystemMssql = "mssql"

	// MySQL
	DbSystemMysql = "mysql"

	// Oracle Database
	DbSystemOracle = "oracle"

	// IBM Db2
	DbSystemDb2 = "db2"

	// PostgreSQL
	DbSystemPostgresql = "postgresql"

	// Amazon Redshift
	DbSystemRedshift = "redshift"

	// Apache Hive
	DbSystemHive = "hive"

	// Cloudscape
	DbSystemCloudscape = "cloudscape"

	// HyperSQL DataBase
	DbSystemHsqlsb = "hsqlsb"

	// Progress Database
	DbSystemProgress = "progress"

	// SAP MaxDB
	DbSystemMaxdb = "maxdb"

	// SAP HANA
	DbSystemHanadb = "hanadb"

	// Ingres
	DbSystemIngres = "ingres"

	// FirstSQL
	DbSystemFirstsql = "firstsql"

	// EnterpriseDB
	DbSystemEdb = "edb"

	// InterSystems Caché
	DbSystemCache = "cache"

	// Adabas (Adaptable Database System)
	DbSystemAdabas = "adabas"

	// Firebird
	DbSystemFirebird = "firebird"

	// Apache Derby
	DbSystemDerby = "derby"

	// FileMaker
	DbSystemFilemaker = "filemaker"

	// Informix
	DbSystemInformix = "informix"

	// InstantDB
	DbSystemInstantdb = "instantdb"

	// InterBase
	DbSystemInterbase = "interbase"

	// MariaDB
	DbSystemMariadb = "mariadb"

	// Netezza
	DbSystemNetezza = "netezza"

	// Pervasive PSQL
	DbSystemPervasive = "pervasive"

	// PointBase
	DbSystemPointbase = "pointbase"

	// SQLite
	DbSystemSqlite = "sqlite"

	// Sybase
	DbSystemSybase = "sybase"

	// Teradata
	DbSystemTeradata = "teradata"

	// Vertica
	DbSystemVertica = "vertica"

	// H2
	DbSystemH2 = "h2"

	// ColdFusion IMQ
	DbSystemColdfusion = "coldfusion"

	// Apache Cassandra
	DbSystemCassandra = "cassandra"

	// Apache HBase
	DbSystemHbase = "hbase"

	// MongoDB
	DbSystemMongodb = "mongodb"

	// Redis
	DbSystemRedis = "redis"

	// Couchbase
	DbSystemCouchbase = "couchbase"

	// CouchDB
	DbSystemCouchdb = "couchdb"

	// Microsoft Azure Cosmos DB
	DbSystemCosmosdb = "cosmosdb"

	// Amazon DynamoDB
	DbSystemDynamodb = "dynamodb"

	// Neo4j
	DbSystemNeo4j = "neo4j"
)

// The available values for net.transport.
const (
	// ip_tcp
	NetTransportIpTcp = "IP.TCP"

	// ip_udp
	NetTransportIpUdp = "IP.UDP"

	// Another IP-based protocol
	NetTransportIp = "IP"

	// Unix Domain socket. See below
	NetTransportUnix = "Unix"

	// Named or anonymous pipe. See note below
	NetTransportPipe = "pipe"

	// In-process communication
	NetTransportInproc = "inproc"

	// Something else (non IP-based)
	NetTransportOther = "other"
)

// The available values for dt.ctg.requesttype.
const (
	// Base. A base GatewayRequest without a further subtype
	DtCtgRequesttypeBase = "BASE"

	// External Call Interface. Enables a client application to call a CICS program synchronously or asynchronously
	DtCtgRequesttypeEci = "ECI"

	// External Presentation Interface. Enables a user application to install a virtual IBM 3270 terminal into a CICS server
	DtCtgRequesttypeEpi = "EPI"

	// External Security Interface. Enables user applications to perform security-related tasks
	DtCtgRequesttypeEsi = "ESI"

	// CICS Request Exit. Can be used for request retry, dynamic server selection and for rejecting non-valid requests
	DtCtgRequesttypeXa = "XA"

	// Admin
	DtCtgRequesttypeAdmin = "ADMIN"

	// Authentication
	DtCtgRequesttypeAuth = "AUTH"
)

// The available values for dt.ctg.calltype.
const (
	// eci_unknown_call_type
	DtCtgCalltypeEciUnknownCallType = "0"

	// eci_sync
	DtCtgCalltypeEciSync = "1"

	// eci_async
	DtCtgCalltypeEciAsync = "2"

	// eci_get_reply
	DtCtgCalltypeEciGetReply = "3"

	// eci_get_reply_wait
	DtCtgCalltypeEciGetReplyWait = "4"

	// eci_get_specific_reply
	DtCtgCalltypeEciGetSpecificReply = "5"

	// eci_get_specific_reply_wait
	DtCtgCalltypeEciGetSpecificReplyWait = "6"

	// eci_state_sync
	DtCtgCalltypeEciStateSync = "7"

	// eci_state_async
	DtCtgCalltypeEciStateAsync = "8"

	// cics_eci_list_systems
	DtCtgCalltypeCicsEciListSystems = "9"

	// eci_state_sync_java
	DtCtgCalltypeEciStateSyncJava = "10"

	// eci_state_async_java
	DtCtgCalltypeEciStateAsyncJava = "11"

	// eci_sync_tpn
	DtCtgCalltypeEciSyncTpn = "12"

	// eci_async_tpn
	DtCtgCalltypeEciAsyncTpn = "13"
)

// The available values for dt.db.topology.
const (
	// not_set
	DtDbTopologyNotSet = "not_set"

	// single_server
	DtDbTopologySingleServer = "single_server"

	// embedded
	DtDbTopologyEmbedded = "embedded"

	// failover
	DtDbTopologyFailover = "failover"

	// load_balancing
	DtDbTopologyLoadBalancing = "load_balancing"

	// local_ipc
	DtDbTopologyLocalIpc = "local_ipc"

	// cluster
	DtDbTopologyCluster = "cluster"
)

// The available values for faas.trigger.
const (
	// A response to some data source operation such as a database or filesystem read/write
	FaasTriggerDatasource = "datasource"

	// To provide an answer to an inbound HTTP request
	FaasTriggerHttp = "http"

	// A function is set to be executed when messages are sent to a messaging system
	FaasTriggerPubsub = "pubsub"

	// A function is scheduled to be executed regularly
	FaasTriggerTimer = "timer"

	// If none of the others apply
	FaasTriggerOther = "other"
)

// The available values for http.flavor.
const (
	// HTTP 1.0
	HttpFlavorHttp10 = "1.0"

	// HTTP 1.1
	HttpFlavorHttp11 = "1.1"

	// HTTP 2
	HttpFlavorHttp20 = "2.0"

	// SPDY protocol
	HttpFlavorSpdy = "SPDY"

	// QUIC protocol
	HttpFlavorQuic = "QUIC"
)

// The available values for messaging.destination_kind.
const (
	// queue
	MessagingDestinationKindQueue = "queue"

	// topic
	MessagingDestinationKindTopic = "topic"
)

// The available values for dt.messaging.jms.message_type.
const (
	// other
	DtMessagingJmsMessageTypeOther = "other"

	// map
	DtMessagingJmsMessageTypeMap = "map"

	// object
	DtMessagingJmsMessageTypeObject = "object"

	// stream
	DtMessagingJmsMessageTypeStream = "stream"

	// bytes
	DtMessagingJmsMessageTypeBytes = "bytes"

	// text
	DtMessagingJmsMessageTypeText = "text"
)

// The available values for dt.zosconnect.sor_type.
const (
	// cics
	DtZosconnectSorTypeCics = "CICS"

	// ims
	DtZosconnectSorTypeIms = "IMS"

	// rest
	DtZosconnectSorTypeRest = "REST"

	// wola
	DtZosconnectSorTypeWola = "WOLA"

	// mq
	DtZosconnectSorTypeMq = "MQ"
)

// The available values for dt.zosconnect.request_type.
const (
	// api
	DtZosconnectRequestTypeApi = "API"

	// service
	DtZosconnectRequestTypeService = "SERVICE"

	// admin
	DtZosconnectRequestTypeAdmin = "ADMIN"

	// unknown
	DtZosconnectRequestTypeUnknown = "UNKNOWN"
)

// The available values for faas.document.operation.
const (
	// When a new object is created
	FaasDocumentOperationInsert = "insert"

	// When an object is modified
	FaasDocumentOperationEdit = "edit"

	// When an object is deleted
	FaasDocumentOperationDelete = "delete"
)

// The available values for faas.invoked_provider.
const (
	// Amazon Web Services
	FaasInvokedProviderAws = "aws"

	// Microsoft Azure
	FaasInvokedProviderAzure = "azure"

	// Google Cloud Platform
	FaasInvokedProviderGcp = "gcp"
)

// The available values for rpc.system.
const (
	// gRPC
	RpcSystemGrpc = "grpc"

	// Java RMI
	RpcSystemJavaRmi = "java_rmi"

	// .NET WCF
	RpcSystemDotnetWcf = "dotnet_wcf"

	// Apache Dubbo
	RpcSystemApacheDubbo = "apache_dubbo"
)

// The available values for messaging.operation.
const (
	// receive
	MessagingOperationReceive = "receive"

	// process
	MessagingOperationProcess = "process"
)

// The available values for rpc.grpc.status_code.
const (
	// OK
	RpcGrpcStatusCodeOk = "0"

	// CANCELLED
	RpcGrpcStatusCodeCancelled = "1"

	// UNKNOWN
	RpcGrpcStatusCodeUnknown = "2"

	// INVALID_ARGUMENT
	RpcGrpcStatusCodeInvalidArgument = "3"

	// DEADLINE_EXCEEDED
	RpcGrpcStatusCodeDeadlineExceeded = "4"

	// NOT_FOUND
	RpcGrpcStatusCodeNotFound = "5"

	// ALREADY_EXISTS
	RpcGrpcStatusCodeAlreadyExists = "6"

	// PERMISSION_DENIED
	RpcGrpcStatusCodePermissionDenied = "7"

	// RESOURCE_EXHAUSTED
	RpcGrpcStatusCodeResourceExhausted = "8"

	// FAILED_PRECONDITION
	RpcGrpcStatusCodeFailedPrecondition = "9"

	// ABORTED
	RpcGrpcStatusCodeAborted = "10"

	// OUT_OF_RANGE
	RpcGrpcStatusCodeOutOfRange = "11"

	// UNIMPLEMENTED
	RpcGrpcStatusCodeUnimplemented = "12"

	// INTERNAL
	RpcGrpcStatusCodeInternal = "13"

	// UNAVAILABLE
	RpcGrpcStatusCodeUnavailable = "14"

	// DATA_LOSS
	RpcGrpcStatusCodeDataLoss = "15"

	// UNAUTHENTICATED
	RpcGrpcStatusCodeUnauthenticated = "16"
)

// The available values for message.type.
const (
	// sent
	MessageTypeSent = "SENT"

	// received
	MessageTypeReceived = "RECEIVED"
)
