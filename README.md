# OpenTelemetry Go exporter for Dynatrace

This is the home of the OpenTelemetry Go trace exporter for Dynatrace. This package allows you to export spans to Dynatrace
using the OpenTelemetry API.

This package is provided as a "core package", meaning that no platform-specific span attributes are automatically
provided, which might be required by Dynatrace in order to enable service detection and correlation.

Please refer to the [Dynatrace documentation](https://www.dynatrace.com/support/help/setup-and-configuration/setup-on-cloud-platforms/google-cloud-platform/opentelemetry-integration/opentelemetry-on-gcf-go) for how to set up tracing for Google Cloud Functions with Go using this package.

## Prerequisites

You need a Dynatrace environment to export your spans to.

We currently only offer support and documentation for using this package on Google Cloud Functions (GCF). For GCF-specific prerequisites, please refer to [Prerequisites](https://www.dynatrace.com/support/help/setup-and-configuration/setup-on-cloud-platforms/google-cloud-platform/opentelemetry-integration/opentelemetry-on-gcf-go#prerequisites).

## Getting started

This README file just contains a very brief quickstart code sample. The main documentation of this package can be found in the [Dynatrace Documentation](https://www.dynatrace.com/support/help/setup-and-configuration/setup-on-cloud-platforms/google-cloud-platform/opentelemetry-integration/opentelemetry-on-gcf-go).

Follow [the instructions](https://www.dynatrace.com/support/help/setup-and-configuration/setup-on-cloud-platforms/google-cloud-platform/opentelemetry-integration/opentelemetry-on-gcf#choose-config-method) on how to set up the Dynatrace configuration for your project by using a config file or environment variables. An error will be returned during the TracerProvider instantiation if the config cannot be found.

1. Import the package into your Go project:
    ```shell
    go get github.com/dynatrace-oss/opentelemetry-exporter-go/core
    ```
2. Create a Dynatrace TracerProvider and TextMapPropagator:
    ```go
    import (
        "go.opentelemetry.io/otel"
        dtTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/trace"
    )
   
    // create a TracerProvider
    tracerProvider, err := dtTrace.NewTracerProvider()
    // optionally handle error
    otel.SetTracerProvider(tracerProvider)

    // create a TextMapPropagator
    propagator, err := dtTrace.NewTextMapPropagator()
    // optionally handle error
    otel.SetTextMapPropagator(prop)
    ```
3. Use the OpenTelemetry API as you normally would to create spans:
    ```go
    // otel.GetTracerProvider() will now return the DtTracerProvider that was created previously.
    tracer := otel.GetTracerProvider().Tracer("example tracer")
    ctx, span := tracer.Start(context.Background(), "example span")
    // do something
    span.End()
    ```

## Support

Before creating a support ticket, please read through the [documentation](https://www.dynatrace.com/support/help/setup-and-configuration/setup-on-cloud-platforms/google-cloud-platform/opentelemetry-integration/opentelemetry-on-gcf-go).

If you didn't find a solution there,
please [contact Dynatrace support](https://www.dynatrace.com/support/contact-support/).

## License

[Apache License Version 2.0](LICENSE)
