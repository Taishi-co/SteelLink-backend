# Spec: AI Procurement Order Grouping And Logistics Visibility Fix

## Background

AI procurement can select multiple matched products and generate real trade orders. A recent production case selected two products at one ton each, but the flow created two order numbers instead of one expected combined order. The logistics workspace also only showed one related logistics order, leaving the other trade order invisible to the carrier-side confirmation flow.

## Requirements

- AI procurement order grouping must use canonical product seller and origin warehouse data from the product catalog, not only the selected-product JSON snapshot returned by earlier matching steps.
- Same-seller products from the same origin warehouse should be generated as one trade order.
- Products from different sellers or different origin warehouses must remain split before calling the order service.
- When AI procurement creates trade orders with a matched logistics service, the backend must create matching `logistics_orders` records for every generated trade order.
- Negotiated logistics or no-match logistics states must not create a carrier-specific logistics order because no concrete logistics service has been selected.

## Impact

- Affects `POST /api/v1/ai-procurements/generate-order`.
- Writes new rows to `logistics_orders` when selected products include `logistics_service_id`.
- Does not change the standalone `POST /api/v1/logistics/orders` API.
- Does not change order status transitions or warehouse confirmation requirements.

## Rollback

Revert the AI procurement service changes. Existing generated trade orders and logistics orders remain as business data and should be handled by operational repair if needed.
