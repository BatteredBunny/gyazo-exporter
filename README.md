# Gyazo Exporter

## Usage

1. Activate free trial or pay for a month of premium
2. [Create new application](https://gyazo.com/oauth/applications)
3. Input name: gyazo-exporter, callback url: http://example.com
4. Click on the newly made application then "Generate" and copy the text
5. ``go run . --access_token XXXX``

## Run with nix flakes

```
nix run github:BatteredBunny/gyazo-exporter
```