# Security Policy

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

To report a security vulnerability, please email: **support@duplocloud.com**

Include the following information:
- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested fixes (optional)

### Response Timeline

- We will acknowledge your report within **3 business days**
- We will provide updates every **7 days**
- We aim to release fixes within **90 days**

## Security Best Practices

When using this provider:

1. **Protect Credentials**
   - Never commit `duplo_token` to version control
   - Use environment variables or secure secret management
   - Rotate tokens regularly

2. **Secure Terraform State**
   - Store state files in encrypted backends (S3, Terraform Cloud, etc.)
   - Never commit state files to version control
   - State files contain sensitive data - restrict access

3. **Use HTTPS**
   - Always use HTTPS for `duplo_host`
   - Only disable `ssl_no_verify` in development environments

4. **Keep Updated**
   - Regularly update to the latest provider version
   - Monitor security advisories

## Contact

- **Security Issues**: support@duplocloud.com
- **General Support**: support@duplocloud.com
- **Non-Security Bugs**: [GitHub Issues](https://github.com/duplocloud/terraform-provider-duplocloud/issues)

---

**Last Updated**: January 22, 2026