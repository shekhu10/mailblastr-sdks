namespace Mailblastr;

/// <summary>
/// Thrown when the MailBlastr API returns a non-2xx response (or the response
/// cannot be reached/parsed). Carries the API error shape:
/// <c>{ statusCode, name, message }</c>.
/// </summary>
public class MailblastrException : Exception
{
    /// <summary>HTTP status code of the failed response (0 for network errors).</summary>
    public int StatusCode { get; }

    /// <summary>Machine-readable error name, e.g. <c>validation_error</c> or <c>not_found</c>.</summary>
    public string Name { get; }

    public MailblastrException(int statusCode, string name, string message)
        : base(message)
    {
        StatusCode = statusCode;
        Name = name;
    }

    public MailblastrException(int statusCode, string name, string message, Exception innerException)
        : base(message, innerException)
    {
        StatusCode = statusCode;
        Name = name;
    }
}
