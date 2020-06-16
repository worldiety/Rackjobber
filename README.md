# Rackjobber

Rackjobber is a CL-Tool written in Go, which automates versioning of Plugins and Themes for Shopware projects. 
Use the 'up' command to update and install all configured Plugins and Themes for your specified Shopware project


## Installation

Refer to the release Page for archives or use Brew:
```
brew tap wdy/rackjobber https://github.com/worldiety/rackjobber
brew install wdy/rackjobber/rackjobber
rackjobber
```

## Considering Themes used by Rackjobber:

To set a Theme shopware uses the namespace specified in the `Theme.php`, which is stored in `[PluginName]/Resources/Themes/Frontend/[ThemeName]/Theme.php`
Rackjobber is only able to set a Theme, when the corresponding `rackspec.yaml` specifies this namespace in the `theme: ` section

### If shopware does not display your theme correctly, consider the following possible solutions:

**Does the Theme Manager in shopware recognize your chosen theme?**


To check this, access the Backend of your Shopware store at `[shopdomain]/backend/`, and open `Configuration -> Theme Manager`. 
If your chosen Theme is not listed here, rackjobber was not able to integrate it into shopware. Please check the following:

1. Is your Theme listed in the `rackfile.yaml` on your shopware server?
2. Is your Theme's `rackspec.yaml` up to date, as well as the specified Git-URL in it? 

**shopware recognizes my Theme, but Rackjobber's up does not set it**


Please check the following:
1. Does the `rackfile.yaml`on your shopware server contain multiple Themes? In this case, Rackjobber will set the latest one of the specified Themes.
2. Does your `rackspec.yaml`'s set Theme correspond to the Theme's namespace? 
   To check this, navigate to the `rackspec.yaml` of the version of your Theme, 
   to be found at `rackjobber/rackressources/repos/master/[ThemeName]/[ThemeVersion]/[ThemeName]_rackspec.yaml` and look for the `theme` section inside this yaml,
   looking something like this: 



```yaml
theme: TestTheme
```
    
    
    
   Now, navigate to your Theme's `Theme.php` on your shopware Server, to be found at `[PluginName]/Resources/Themes/Frontend/[ThemeName]/Theme.php` 
   and look for the `namespace` section, looking something like this:
   
   
   
```php
<?php

namespace Shopware\Themes\TestTheme;

use Shopware\Components\Form as Form;
```
    
    

   Here, the Theme specified in the `rackpec.yaml` has to correspond to the last part of the `Theme.php`'s namespace. 
   If it does not, please change the `rackspec.yaml` accordingly.
   
   
   
## SSH Connection

Rackjobber uses the SSH protocoll to establish a secure connection to the shopware server.
It uses the public host key and the private rsa key at their default locations `~/.ssh/known_hosts` and `~/.ssh/id_rsa`
These keys have to be setup manually. Information on how to setup an initial SSH connection can be found [here](https://www.digitalocean.com/community/tutorials/how-to-set-up-ssh-keys--2).
