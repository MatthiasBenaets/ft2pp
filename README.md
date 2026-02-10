# ft2pp

This project provides a way to retrieve Financial Times securities chart data and format it for use with Portfolio Performance.

---

### Usage with Portfolio Performance:

1. Edit a security in Portfolio Performance.
2. Go to "Historical Quotes".
3. Pick "JSON" as the data source.
4. Enter the URL, for example:
   http://localhost:8080/api/market-data?symbol=INX:IOM&id=575769&start=2020-01-01

   (Note: Replace 'INX:IOM', '575769', and the 'start' date with your specific security details.)

5. Configure the JSON paths as follows:
   - Path to Date: $.Dates[*]
   - Path to Close: $.Elements[0].ComponentSeries[3].Values[*]
   - Path to Day's Low: $.Elements[0].ComponentSeries[2].Values[*]
   - Path to Day's High: $.Elements[0].ComponentSeries[1].Values[*]
   - Path to Volume: $.Elements[1].ComponentSeries[0].Values[*]

##### How to get 'symbol' and 'id' for the URL:

- 'symbol': This is typically the stock's ticker or ISIN as it appears on Financial Times.
- 'id': To find the 'id':
  1.  Open your browser's developer tools (usually F12).
  2.  Navigate to the "Network" tab.
  3.  Filter the requests by "XHR".
  4.  Load the chart on the relevant Financial Times security page.
  5.  Look for a request named "series" (or similar) in the network traffic.
  6.  Inspect the "Request" tab of this "series" request to find the 'id', often located within 'elements[0].Symbol'.

### Deploy:

##### Docker Compose:

To run the application and its Redis dependency using Docker Compose, ensure you have Docker and Docker Compose installed, then execute:

```bash
docker-compose up -d
```

##### Running Directly:

1.  **Start a Redis Instance**

2.  **Run the Go Application**
    Navigate to the project root and execute:
    ```bash
    go run .
    ```
    If your Redis instance is not on `localhost:6379`, set the `REDIS_ADDR` environment variable:
    ```bash
    REDIS_ADDR=your_redis_host:port go run .
    ```

### Disclaimers:

- I am not affiliated with Financial Times nor with Portfolio Performance.
- This tool is intended for personal use only.
- An FT developer API key is not required for this tool.
- Please limit calls to once per day to ensure fair use. Requests are aggressively cached.

