# Web Pages

This section describes how the Web UI works and how to use it.

## Home Screen

The home screen is the default landing page for the Web UI. It shows a brief summary of the system status

![Home Page Screenshot](/images/Home_Page.png)

## DHCP

The DHCP page shows settings and current leases for the DHCP server.

{{< figure src="/images/DHCP_Leases.png" caption="Active DHCP Leases" >}}
{{< figure src="/images/DHCP_Reserved_leases.png" caption="Reserved DHCP Leases" >}}

### DHCP Settings

DHCP Settings are saved under the `DHCP` section of the configuration file.

{{< figure src="/images/DHCP_settings.png" caption="DHCP Settings" >}}

Clicking on the Pencil icon next to the Settings will title will open the Settings in a modal dialog.

![DHCP Settings Modal](/images/DHCP_Settings_Edit.png)

## DNS

The DNS page shows local domains, statistics and settings for the DNS server.

### Local Domains

Local domains are saved under the `LocalDomains` section of the configuration file.

{{< figure src="/images/Local_DNS.png" >}}

Clicking the `Add` button will open a modal dialog to add a new local domain, and clicking on an entry in the list will open a modal dialog to edit it.

### DNS Settings

{{< figure src="/images/DNS_Settings.png" >}}

Clicking on the Pencil icon next to the Settings will title will open the Settings in a modal dialog.

![DNS Settings Modal](/images/DNS_Settings_Edit.png)

### DNS Statistics

The DNS Statistics page shows statistics for the DNS server.

![DNS Statistics](/images/DNS_Stats.png)
