Result:
---

I limited time spending on the implementation to several hours as it was suggested in the email.
Which means some parts are still remains unimplemented, but I left my comments with consideration about potential options of implementations.

1. Ad Selection Logic. The most interesting part in terms of storage.
Since the main fetching is 'GetWinningAds' which uses one string 'placement' and two arrays 'categories' and 'keywords'.
There are several way to design a storage for faster fetching 'GetWinningAds'. The first guess is to use PostgreSQL, there is even a index for array columns exists.
Another option could be of using multiple rows with strings of 'category' and 'keyword' instead of arrays. In that case combined index will work better.
Another option is to prepare data for faster load, and cache it for all available combination of 'placement', 'category' and 'keyword'. But it really depends on the average data.
As a summary there are different ways to proceed, but in order to decide we need to see actual average data and the load profile.

2. Tracking Endpoint. I think in terms of data storage the most popular solution for such case is any Message Queue, as example Kafka.
We need to persist data as fast as possible, and then other internal services would process it without races and more resilient for traffic peaks.

3. Relevancy System. That part is the most uncertain. I didn't find any requirements or a product description in terms of what we actually want to take into consideration for relevance calculation.


TODO for production readiness:
* [ ] configure linters (golangci-lint)
* [ ] add metrics (4 golden signals + custom metrics)
* [ ] configure CI/CD
