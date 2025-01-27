# Notes

### Install @kdex-ui from local build

```bash
cd ~/projects/kdex-ui
npm pack

cd ~/projects/kdex-proxy
npm i ../../kdex-ui/kdex-ui-<version>.tgz         
```

### Headers

Sample headers of a request from the browser through the proxy:
```json
{
  "Accept": [
    "*/*"
  ],
  "Accept-Encoding": [
    "gzip, deflate, br, zstd"
  ],
  "Accept-Language": [
    "en-US,en;q=0.5"
  ],
  "Cache-Control": [
    "no-cache"
  ],
  "Cookie": [
    "COOKIE_SUPPORT=true; GUEST_LANGUAGE_ID=en_US; LFR_SESSION_STATE_20096=1737319987644; LFR_SESSION_STATE_20123=1737320495627; JSESSIONID=D7849EDF01BD456F99BE61C2EAFBAF78; COMPANY_ID=90516109914669; ID=6e676f64484c4178446a4d2f747a56672b36617a67773d3d"
  ],
  "Pragma": [
    "no-cache"
  ],
  "Referer": [
    "http://main.dxp.docker.localhost/group/guest/~/control_panel/manage?p_p_id=com_liferay_configuration_admin_web_portlet_SystemSettingsPortlet\u0026p_p_lifecycle=0\u0026p_p_state=maximized\u0026p_p_mode=view\u0026_com_liferay_configuration_admin_web_portlet_SystemSettingsPortlet_factoryPid=com.liferay.commerce.configuration.CommerceOrderConfiguration\u0026_com_liferay_configuration_admin_web_portlet_SystemSettingsPortlet_mvcRenderCommandName=%2Fconfiguration_admin%2Fedit_configuration\u0026_com_liferay_configuration_admin_web_portlet_SystemSettingsPortlet_pid=com.liferay.commerce.configuration.CommerceOrderConfiguration"
  ],
  "Sec-Fetch-Dest": [
    "script"
  ],
  "Sec-Fetch-Mode": [
    "cors"
  ],
  "Sec-Fetch-Site": [
    "same-origin"
  ],
  "User-Agent": [
    "Mozilla/5.0 (X11; Linux x86_64; rv:134.0) Gecko/20100101 Firefox/134.0"
  ],
  "X-Forwarded-For": [
    "10.42.0.1"
  ],
  "X-Forwarded-Host": [
    "main.dxp.docker.localhost"
  ],
  "X-Forwarded-Port": [
    "80"
  ],
  "X-Forwarded-Proto": [
    "http"
  ],
  "X-Forwarded-Server": [
    "traefik-d7c9c5778-nbznj"
  ],
  "X-Real-Ip": [
    "10.42.0.1"
  ]
}
```