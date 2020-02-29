# GetOilPrice Server

This service is created to get updated fuel price in China. It is easily to deploy in any server just like raspberry.

A GET method is provided for querying price with location information.
```
server-address/price?location=your-location
```
Redis and MYSQL is required when try to deploy the service.

Currently it does not support adding your custom price data into database. All data is synchronized in a fixed time
period.

## Shortcuts

An iOS shortcuts can be the front-end of the service. Please visit 
https://shortcuts.sspai.com/#/main/workflow

Shortcuts name: `今日油价`