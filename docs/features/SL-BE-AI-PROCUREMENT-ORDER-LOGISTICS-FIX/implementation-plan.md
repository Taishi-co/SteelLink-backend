# Implementation Plan: AI Procurement Order Grouping And Logistics Visibility Fix

1. Normalize selected AI procurement products from the product catalog before logistics enrichment.
2. Keep order grouping keyed by seller plus canonical warehouse signature.
3. Return order creation plans that preserve the selected products belonging to each real order.
4. Create logistics fulfillment rows for each generated order when the group has a concrete logistics service.
5. Extend focused unit tests for grouping fallback and logistics-order payload helpers.
6. Run focused Go tests and build verification.
