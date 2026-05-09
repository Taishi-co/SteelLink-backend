-- SteelLink demo seed data for steel products, enterprises, logistics, and warehouses.
-- Usage:
--   mysql -h <host> -u <user> -p <database> < scripts/seed-steel-market-demo.sql
--
-- All demo accounts use password: 123456
-- bcrypt hash source: bcrypt("123456")

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

SET @now = NOW();
SET @password_hash = '$2a$10$WLEh/ysKDPKaVu1EzXavDO58NTyui28uuliB/WD8iqtzp30IooF7a';

-- Enterprises
INSERT INTO enterprises (
  id, enterprise_code, name, credit_code, enterprise_role, status, verification_info, remark,
  created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9101, 'DEMO-BUYER-001', '上海钢链汽车零部件有限公司', '91310000DEMO000001', 'buyer', 1,
   JSON_OBJECT('legal_person','周明','legal_id_card','310101198601010011','business_license','91310000DEMO000001','bank_account','622200100000000001','bank_name','招商银行上海分行','contact_name','周明','contact_phone','13800001001','address','上海市嘉定区安亭镇墨玉南路888号'),
   JSON_OBJECT('seed','steel_market_demo','scenario','buyer'), 0, @now, 0, @now, NULL),
  (9102, 'DEMO-BUYER-002', '武汉智造供应链有限公司', '91420100DEMO000002', 'buyer', 1,
   JSON_OBJECT('legal_person','陈雪','legal_id_card','420101198702020022','business_license','91420100DEMO000002','bank_account','622200100000000002','bank_name','中国银行武汉分行','contact_name','陈雪','contact_phone','13800001002','address','湖北省武汉市青山区和平大道999号'),
   JSON_OBJECT('seed','steel_market_demo','scenario','buyer'), 0, @now, 0, @now, NULL),
  (9201, 'DEMO-SELLER-BG', '宝山钢铁股份有限公司', '91310000DEMO100001', 'seller', 1,
   JSON_OBJECT('legal_person','王海','legal_id_card','310101198001010033','business_license','91310000DEMO100001','bank_account','622200100000100001','bank_name','工商银行宝山支行','contact_name','王海','contact_phone','13900002001','address','上海市宝山区富锦路885号'),
   JSON_OBJECT('seed','steel_market_demo','source','ouyeel_sample'), 0, @now, 0, @now, NULL),
  (9202, 'DEMO-SELLER-BM', '上海北铭高强度钢材有限公司', '91310000DEMO100002', 'seller', 1,
   JSON_OBJECT('legal_person','刘强','legal_id_card','310101198203030044','business_license','91310000DEMO100002','bank_account','622200100000100002','bank_name','浦发银行上海分行','contact_name','刘强','contact_phone','13900002002','address','上海市宝山区友谊路1588号'),
   JSON_OBJECT('seed','steel_market_demo','source','ouyeel_sample'), 0, @now, 0, @now, NULL),
  (9203, 'DEMO-SELLER-ZJ', '湛江宝钢高强钢科技有限公司', '91440800DEMO100003', 'seller', 1,
   JSON_OBJECT('legal_person','林涛','legal_id_card','440801198404040055','business_license','91440800DEMO100003','bank_account','622200100000100003','bank_name','建设银行湛江分行','contact_name','林涛','contact_phone','13900002003','address','广东省湛江市东海岛钢铁基地'),
   JSON_OBJECT('seed','steel_market_demo','source','ouyeel_sample'), 0, @now, 0, @now, NULL),
  (9301, 'DEMO-WH-SH-BG116', '上海宝钢运输有限公司116号库运营部', '91310000DEMO200001', 'warehouse', 1,
   JSON_OBJECT('legal_person','赵磊','legal_id_card','310101198505050066','business_license','91310000DEMO200001','bank_account','622200100000200001','bank_name','交通银行宝山支行','contact_name','赵磊','contact_phone','13700003001','address','上海市宝山区蕰川路116号库'),
   JSON_OBJECT('seed','steel_market_demo','warehouse_code','WS005593'), 0, @now, 0, @now, NULL),
  (9302, 'DEMO-WH-SH-TZ75', '上海宝钢运输有限公司75号库运营部', '91310000DEMO200002', 'warehouse', 1,
   JSON_OBJECT('legal_person','孙洁','legal_id_card','310101198606060077','business_license','91310000DEMO200002','bank_account','622200100000200002','bank_name','上海银行宝山支行','contact_name','孙洁','contact_phone','13700003002','address','上海市宝山区谭庄路75号库'),
   JSON_OBJECT('seed','steel_market_demo','warehouse_code','WS004693'), 0, @now, 0, @now, NULL),
  (9303, 'DEMO-WH-NJ-MS', '梅盛华康库运营有限公司', '91320100DEMO200003', 'warehouse', 1,
   JSON_OBJECT('legal_person','钱峰','legal_id_card','320101198707070088','business_license','91320100DEMO200003','bank_account','622200100000200003','bank_name','南京银行江北支行','contact_name','钱峰','contact_phone','13700003003','address','江苏省南京市浦口区梅盛华康库'),
   JSON_OBJECT('seed','steel_market_demo','warehouse_code','WS000825'), 0, @now, 0, @now, NULL),
  (9401, 'DEMO-LOG-SH-HUAYUN', '上海华运钢链物流有限公司', '91310000DEMO300001', 'logistics', 1,
   JSON_OBJECT('legal_person','马亮','legal_id_card','310101198808080099','business_license','91310000DEMO300001','bank_account','622200100000300001','bank_name','民生银行上海分行','contact_name','马亮','contact_phone','13600004001','address','上海市宝山区铁山路88号'),
   JSON_OBJECT('seed','steel_market_demo','fleet_size',46), 0, @now, 0, @now, NULL),
  (9402, 'DEMO-LOG-NJ-SUDA', '南京速达钢材运输有限公司', '91320100DEMO300002', 'logistics', 1,
   JSON_OBJECT('legal_person','吴凯','legal_id_card','320101198909090100','business_license','91320100DEMO300002','bank_account','622200100000300002','bank_name','南京银行浦口支行','contact_name','吴凯','contact_phone','13600004002','address','江苏省南京市浦口区桥林物流园'),
   JSON_OBJECT('seed','steel_market_demo','fleet_size',32), 0, @now, 0, @now, NULL),
  (9403, 'DEMO-LOG-ZJ-GANGHANG', '湛江港航大宗物流有限公司', '91440800DEMO300003', 'logistics', 1,
   JSON_OBJECT('legal_person','黄睿','legal_id_card','440801199001010111','business_license','91440800DEMO300003','bank_account','622200100000300003','bank_name','广发银行湛江分行','contact_name','黄睿','contact_phone','13600004003','address','广东省湛江市霞山区港务大道88号'),
   JSON_OBJECT('seed','steel_market_demo','fleet_size',28), 0, @now, 0, @now, NULL)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  enterprise_role = VALUES(enterprise_role),
  status = VALUES(status),
  verification_info = VALUES(verification_info),
  remark = VALUES(remark),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Main enterprise users. All passwords are "123456".
INSERT INTO users (
  id, uid, user_no, username, platform_id, personal_code, enterprise_code, role, phone, qq, email,
  password, real_name, id_card, identity_status, parent_id, permissions, status, remark,
  created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9111, 'UIDDEMO9101', 'UDEMO9101', 'buyer_shanghai', '1', 'PDEMO9101', 'DEMO-BUYER-001', 'enterprise', '13800001001', '', 'buyer_shanghai@demo.steellink.local',
   @password_hash, '周明', '310101198601010011', 2, 0, JSON_ARRAY('order:create','order:read','procurement:use'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9112, 'UIDDEMO9102', 'UDEMO9102', 'buyer_wuhan', '1', 'PDEMO9102', 'DEMO-BUYER-002', 'enterprise', '13800001002', '', 'buyer_wuhan@demo.steellink.local',
   @password_hash, '陈雪', '420101198702020022', 2, 0, JSON_ARRAY('order:create','order:read','procurement:use'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9211, 'UIDDEMO9201', 'UDEMO9201', 'seller_baosteel', '1', 'PDEMO9201', 'DEMO-SELLER-BG', 'enterprise', '13900002001', '', 'seller_baosteel@demo.steellink.local',
   @password_hash, '王海', '310101198001010033', 2, 0, JSON_ARRAY('product:create','product:read','order:read'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9212, 'UIDDEMO9202', 'UDEMO9202', 'seller_beiming', '1', 'PDEMO9202', 'DEMO-SELLER-BM', 'enterprise', '13900002002', '', 'seller_beiming@demo.steellink.local',
   @password_hash, '刘强', '310101198203030044', 2, 0, JSON_ARRAY('product:create','product:read','order:read'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9213, 'UIDDEMO9203', 'UDEMO9203', 'seller_zhanjiang', '1', 'PDEMO9203', 'DEMO-SELLER-ZJ', 'enterprise', '13900002003', '', 'seller_zhanjiang@demo.steellink.local',
   @password_hash, '林涛', '440801198404040055', 2, 0, JSON_ARRAY('product:create','product:read','order:read'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9311, 'UIDDEMO9301', 'UDEMO9301', 'warehouse_bg116', '1', 'PDEMO9301', 'DEMO-WH-SH-BG116', 'enterprise', '13700003001', '', 'warehouse_bg116@demo.steellink.local',
   @password_hash, '赵磊', '310101198505050066', 2, 0, JSON_ARRAY('warehouse:manage','warehouse:movement'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9312, 'UIDDEMO9302', 'UDEMO9302', 'warehouse_tz75', '1', 'PDEMO9302', 'DEMO-WH-SH-TZ75', 'enterprise', '13700003002', '', 'warehouse_tz75@demo.steellink.local',
   @password_hash, '孙洁', '310101198606060077', 2, 0, JSON_ARRAY('warehouse:manage','warehouse:movement'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9313, 'UIDDEMO9303', 'UDEMO9303', 'warehouse_meisheng', '1', 'PDEMO9303', 'DEMO-WH-NJ-MS', 'enterprise', '13700003003', '', 'warehouse_meisheng@demo.steellink.local',
   @password_hash, '钱峰', '320101198707070088', 2, 0, JSON_ARRAY('warehouse:manage','warehouse:movement'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9411, 'UIDDEMO9401', 'UDEMO9401', 'logistics_huayun', '1', 'PDEMO9401', 'DEMO-LOG-SH-HUAYUN', 'enterprise', '13600004001', '', 'logistics_huayun@demo.steellink.local',
   @password_hash, '马亮', '310101198808080099', 2, 0, JSON_ARRAY('logistics:manage','logistics:quote'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9412, 'UIDDEMO9402', 'UDEMO9402', 'logistics_suda', '1', 'PDEMO9402', 'DEMO-LOG-NJ-SUDA', 'enterprise', '13600004002', '', 'logistics_suda@demo.steellink.local',
   @password_hash, '吴凯', '320101198909090100', 2, 0, JSON_ARRAY('logistics:manage','logistics:quote'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL),
  (9413, 'UIDDEMO9403', 'UDEMO9403', 'logistics_ganghang', '1', 'PDEMO9403', 'DEMO-LOG-ZJ-GANGHANG', 'enterprise', '13600004003', '', 'logistics_ganghang@demo.steellink.local',
   @password_hash, '黄睿', '440801199001010111', 2, 0, JSON_ARRAY('logistics:manage','logistics:quote'), 1, JSON_OBJECT('seed','steel_market_demo'), 0, @now, 0, @now, NULL)
ON DUPLICATE KEY UPDATE
  username = VALUES(username),
  enterprise_code = VALUES(enterprise_code),
  role = VALUES(role),
  password = VALUES(password),
  real_name = VALUES(real_name),
  identity_status = VALUES(identity_status),
  permissions = VALUES(permissions),
  status = VALUES(status),
  remark = VALUES(remark),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Service permissions
INSERT INTO enterprise_trade_capability_permissions (
  enterprise_id, enterprise_code, service, status, audit_reason, approved_at, created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9101, 'DEMO-BUYER-001', 'procurement', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9102, 'DEMO-BUYER-002', 'procurement', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9201, 'DEMO-SELLER-BG', 'sales', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9202, 'DEMO-SELLER-BM', 'sales', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9203, 'DEMO-SELLER-ZJ', 'sales', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL)
ON DUPLICATE KEY UPDATE status = VALUES(status), approved_at = VALUES(approved_at), updated_at = VALUES(updated_at), deleted_at = NULL;

INSERT INTO enterprise_provider_service_permissions (
  enterprise_id, enterprise_code, service, status, audit_reason, approved_at, created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9301, 'DEMO-WH-SH-BG116', 'warehouse', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9302, 'DEMO-WH-SH-TZ75', 'warehouse', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9303, 'DEMO-WH-NJ-MS', 'warehouse', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9401, 'DEMO-LOG-SH-HUAYUN', 'logistics', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9402, 'DEMO-LOG-NJ-SUDA', 'logistics', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL),
  (9403, 'DEMO-LOG-ZJ-GANGHANG', 'logistics', 2, 'demo seed approved', @now, 0, @now, 0, @now, NULL)
ON DUPLICATE KEY UPDATE status = VALUES(status), approved_at = VALUES(approved_at), updated_at = VALUES(updated_at), deleted_at = NULL;

-- Warehouse services
INSERT INTO warehouse_services (
  id, enterprise_id, warehouse_name, province, city, district, address, capacity, available_capacity,
  service_types, price_rule, loading_capacity, business_hours, pledge_supervision, outbound_voucher_req,
  images, certificates, description, status, audit_remark, approved_at, created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9501, 9301, '上海宝钢运输有限公司116号库(洪澳库)', '上海市', '上海市', '宝山区', '上海市宝山区蕰川路116号库', 120000.00, 46500.00,
   JSON_ARRAY('现货托管','出入库作业','质押监管','视频验货'), '仓储费0.42元/吨/天，装卸费18元/吨，月结可议', '单日装卸3500吨，行车32台，支持卷板吊装', '工作日 09:00-11:30, 13:30-16:30', 1, '提货委托书、电子仓单、司机身份证、车牌号',
   JSON_ARRAY('/demo/warehouses/sh-bg-116-main.jpg'), JSON_ARRAY('/demo/certs/sh-bg-116-license.pdf'), '宝钢系上海核心仓，适合冷轧、镀锌、电镀锌板卷托管。', 2, 'demo seed approved', @now, 9311, @now, 9311, @now, NULL),
  (9502, 9302, '上海宝钢运输有限公司75号库(谭庄库)', '上海市', '上海市', '宝山区', '上海市宝山区谭庄路75号库', 90000.00, 38200.00,
   JSON_ARRAY('现货托管','剪切分拨','装车预约'), '仓储费0.39元/吨/天，装卸费16元/吨', '单日装卸2800吨，支持整卷与零卷分拨', '工作日 09:00-11:30, 13:30-16:30', 1, '电子提单、出库预约单、承运车辆备案',
   JSON_ARRAY('/demo/warehouses/sh-tz-75-main.jpg'), JSON_ARRAY('/demo/certs/sh-tz-75-license.pdf'), '上海区域镀锌、电镀锌与汽车板卷分拨仓。', 2, 'demo seed approved', @now, 9312, @now, 9312, @now, NULL),
  (9503, 9303, '梅盛华康库', '江苏省', '南京市', '浦口区', '江苏省南京市浦口区梅盛华康库', 110000.00, 52400.00,
   JSON_ARRAY('现货托管','热轧卷堆存','跨城调拨'), '仓储费0.35元/吨/天，装卸费15元/吨', '单日装卸3200吨，适配热轧直发卷', '每日 08:30-17:30', 0, '提货函、车队派车单、司机实名信息',
   JSON_ARRAY('/demo/warehouses/nj-ms-main.jpg'), JSON_ARRAY('/demo/certs/nj-ms-license.pdf'), '南京热轧资源核心仓，支持华东区域汽车、机械客户快速调拨。', 2, 'demo seed approved', @now, 9313, @now, 9313, @now, NULL)
ON DUPLICATE KEY UPDATE
  enterprise_id = VALUES(enterprise_id),
  warehouse_name = VALUES(warehouse_name),
  available_capacity = VALUES(available_capacity),
  service_types = VALUES(service_types),
  price_rule = VALUES(price_rule),
  status = VALUES(status),
  approved_at = VALUES(approved_at),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Logistics services
INSERT INTO logistics_services (
  id, enterprise_id, name, route_start, route_end, service_scope, vehicle_type, load_capacity, price_rule, quote_rules,
  fleet_name, vehicle_certs, driver_certs, images, certificates, description, status, audit_remark, approved_at,
  created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9601, 9401, '上海宝山/嘉定钢卷短驳专线', '上海市宝山区', '上海市嘉定区', '宝山、嘉定、青浦、松江钢卷短驳与工厂直送', '13米半挂/卷钢鞍座车', 32.00,
   '基础价900元/车，20公里后8.5元/公里，夜间加收10%', JSON_OBJECT('base_fee',900,'base_km',20,'per_km',8.5,'min_weight_ton',5,'max_weight_ton',32,'night_surcharge_rate',0.10),
   '华运宝山车队', JSON_ARRAY('/demo/certs/vehicle-sh-001.pdf'), JSON_ARRAY('/demo/certs/driver-sh-001.pdf'), JSON_ARRAY('/demo/logistics/sh-huayun-main.jpg'), JSON_ARRAY('/demo/certs/log-sh-license.pdf'),
   '服务上海宝钢仓库至周边汽车零部件、家电和机械加工厂。', 2, 'demo seed approved', @now, 9411, @now, 9411, @now, NULL),
  (9602, 9402, '南京梅盛库华东干线', '南京市浦口区', '上海市/苏州市/无锡市', '南京至华东主要制造城市热轧卷干线运输', '17.5米平板/13米高栏', 35.00,
   '南京至上海180元/吨，南京至苏州155元/吨，南京至无锡145元/吨，整车议价', JSON_OBJECT('price_per_ton_to_shanghai',180,'price_per_ton_to_suzhou',155,'price_per_ton_to_wuxi',145,'min_order_ton',10),
   '速达华东车队', JSON_ARRAY('/demo/certs/vehicle-nj-001.pdf'), JSON_ARRAY('/demo/certs/driver-nj-001.pdf'), JSON_ARRAY('/demo/logistics/nj-suda-main.jpg'), JSON_ARRAY('/demo/certs/log-nj-license.pdf'),
   '覆盖梅盛华康库、梅盛中储库出库后的华东干线运输。', 2, 'demo seed approved', @now, 9412, @now, 9412, @now, NULL),
  (9603, 9403, '湛江钢厂华南干线', '湛江市东海岛', '广州市/佛山市/深圳市', '湛江钢厂资源至珠三角制造企业干线配送', '13米半挂/卷钢专用车', 33.00,
   '湛江至广州210元/吨，湛江至佛山220元/吨，湛江至深圳245元/吨', JSON_OBJECT('price_per_ton_to_guangzhou',210,'price_per_ton_to_foshan',220,'price_per_ton_to_shenzhen',245,'min_order_ton',8),
   '港航华南车队', JSON_ARRAY('/demo/certs/vehicle-zj-001.pdf'), JSON_ARRAY('/demo/certs/driver-zj-001.pdf'), JSON_ARRAY('/demo/logistics/zj-ganghang-main.jpg'), JSON_ARRAY('/demo/certs/log-zj-license.pdf'),
   '适配湛江热镀锌、酸洗、热轧直发卷资源的跨城运输。', 2, 'demo seed approved', @now, 9413, @now, 9413, @now, NULL)
ON DUPLICATE KEY UPDATE
  enterprise_id = VALUES(enterprise_id),
  name = VALUES(name),
  route_start = VALUES(route_start),
  route_end = VALUES(route_end),
  price_rule = VALUES(price_rule),
  quote_rules = VALUES(quote_rules),
  status = VALUES(status),
  approved_at = VALUES(approved_at),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Products. Status 2 means on sale.
INSERT INTO products (
  id, product_no, name, goods_type, category, sub_category, steel_mill, spec, warehouse_service_id, warehouse_enterprise_id,
  warehouse_info, bundle_no, quantity, available_qty, locked_qty, unit_price, min_order_qty, currency, tax_included, tax_rate,
  status, enterprise_id, audit_remark, approved_at, quality_cert, images, description, delivery_days, quote_validity, remark,
  created_by, created_at, updated_by, updated_at, deleted_at
) VALUES
  (9701, 'DEMO-P-9701', '冷轧板卷 HC340LA 2.00*1488*C', 'spot', '冷轧', '冷轧板卷', '宝钢股份',
   JSON_OBJECT('material','HC340LA','spec','2.00*1488*C','thickness','2.00','width','1488','length','C','surface','FB/D','edge','切边','processing','简包装'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-L6W0029699', 20.33, 20.33, 0.00, 4326.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/L6W0029699.pdf', JSON_ARRAY('/demo/products/cold-rolled-coil.jpg'), '取自欧冶样例资源，正品，含仓库认证库存。', 2, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031206237,'provider_name','宝山钢铁股份有限公司','publish_date','2026-05-08 12:13:36'), 9211, @now, 9211, @now, NULL),
  (9702, 'DEMO-P-9702', '热镀锌板卷 DC54D+ZF 0.65*1365*C', 'spot', '热镀锌', '热镀锌板卷', '湛江钢铁',
   JSON_OBJECT('material','DC54D+ZF','spec','0.65*1365*C','thickness','0.65','width','1365','length','C','surface','45/45 无机固体润滑+涂油','edge','切边','processing','锌铁(ZF)'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-M6W0064642', 4.12, 4.12, 0.00, 4176.00, 1.00, 'CNY', 1, 13.00, 2, 9203, 'demo seed approved', @now,
   '/demo/certs/quality/M6W0064642.pdf', JSON_ARRAY('/demo/products/galvanized-coil.jpg'), '湛江钢铁热镀锌现货，小吨位试单资源。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031206563,'provider_name','宝山钢铁股份有限公司'), 9213, @now, 9213, @now, NULL),
  (9703, 'DEMO-P-9703', '热轧板卷 Q355ME 2*1500*C', 'spot', '热轧', '热轧板卷', '本钢',
   JSON_OBJECT('material','Q355ME','spec','2*1500*C','thickness','2.00','width','1500','length','C','surface','','edge','','processing','简包装'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-Q355ME-1500', 203.39, 203.39, 0.00, 4226.00, 5.00, 'CNY', 1, 13.00, 2, 9202, 'demo seed approved', @now,
   '/demo/certs/quality/Q355ME-1500.pdf', JSON_ARRAY('/demo/products/hot-rolled-coil.jpg'), '上海北铭高强度热轧板卷资源，适合结构件采购。', 2, 48,
   JSON_OBJECT('source','ouyeel_sample','sku_id',1451472987818581909,'total_count',9), 9212, @now, 9212, @now, NULL),
  (9704, 'DEMO-P-9704', '热镀锌板卷 DC51D+Z 0.65*1600*C', 'spot', '热镀锌', '热镀锌板卷', '宝钢股份',
   JSON_OBJECT('material','DC51D+Z','spec','0.65*1600*C','thickness','0.65','width','1600','length','C','surface','涂油','edge','切边','processing','纯锌(Z)'),
   9502, 9302, JSON_OBJECT('warehouse_id',9502,'warehouse_service_id',9502,'warehouse_enterprise_id',9302,'warehouse_name','上海宝钢运输有限公司75号库(谭庄库)','warehouse_code','WS004693','address','上海市宝山区谭庄路75号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-TL21BG1001576', 22.98, 22.98, 0.00, 4246.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/TL21BG1001576.pdf', JSON_ARRAY('/demo/products/galvanized-coil.jpg'), '平台好货标签资源，上海库现货。', 2, 72,
   JSON_OBJECT('source','ouyeel_sample','sku_id',1441160917369771671,'tag','平台好货'), 9211, @now, 9211, @now, NULL),
  (9705, 'DEMO-P-9705', '酸洗板卷 SPHC-MJ 2.00*1500*C', 'spot', '酸洗', '酸洗板卷', '湛江钢铁',
   JSON_OBJECT('material','SPHC-MJ','spec','2.00*1500*C','thickness','2.00','width','1500','length','C','surface','FB','edge','不切边','processing','酸洗正品'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-X6W0013545', 11.88, 11.88, 0.00, 4256.00, 1.00, 'CNY', 1, 13.00, 2, 9203, 'demo seed approved', @now,
   '/demo/certs/quality/X6W0013545.pdf', JSON_ARRAY('/demo/products/pickled-coil.jpg'), '湛江酸洗正品，热销精品样例。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031205072,'tag','热销精品'), 9213, @now, 9213, @now, NULL),
  (9706, 'DEMO-P-9706', '冷轧板卷 BLD 1.00*1075*C', 'spot', '冷轧', '冷轧板卷', '宝钢股份',
   JSON_OBJECT('material','BLD','spec','1.00*1075*C','thickness','1.00','width','1075','length','C','surface','FB/D','edge','切边','processing','简包装'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-L6W0032376', 10.33, 10.33, 0.00, 4226.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/L6W0032376.pdf', JSON_ARRAY('/demo/products/cold-rolled-coil.jpg'), '宝钢冷轧 BLD，小批量可下单。', 2, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031206647), 9211, @now, 9211, @now, NULL),
  (9707, 'DEMO-P-9707', '热镀锌板卷 DC54D+Z 0.75*1796*C', 'spot', '热镀锌', '热镀锌板卷', '湛江钢铁',
   JSON_OBJECT('material','DC54D+Z','spec','0.75*1796*C','thickness','0.75','width','1796','length','C','surface','涂油','edge','切边','processing','纯锌(Z)'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-TL21BZJGT1021968', 35.10, 35.10, 0.00, 4326.00, 2.00, 'CNY', 1, 13.00, 2, 9203, 'demo seed approved', @now,
   '/demo/certs/quality/TL21BZJGT1021968.pdf', JSON_ARRAY('/demo/products/galvanized-coil.jpg'), '湛江热镀锌宽幅资源。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','sku_id',1472309171251603898,'total_count',2), 9213, @now, 9213, @now, NULL),
  (9708, 'DEMO-P-9708', '冷轧轧硬卷 SPCC-1B 0.25*1250*C', 'spot', '轧硬卷', '冷轧轧硬卷', '梅钢',
   JSON_OBJECT('material','SPCC-1B','spec','0.25*1250*C','thickness','0.25','width','1250','length','C','surface','FB/B','edge','不切边','processing','简包装'),
   9503, 9303, JSON_OBJECT('warehouse_id',9503,'warehouse_service_id',9503,'warehouse_enterprise_id',9303,'warehouse_name','梅盛华康库','warehouse_code','WS000825','address','江苏省南京市浦口区梅盛华康库','province','江苏省','city','南京市','district','浦口区'),
   'OUYEEL-L6WR002079', 11.41, 11.41, 0.00, 4136.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/L6WR002079.pdf', JSON_ARRAY('/demo/products/hard-rolled-coil.jpg'), '梅钢轧硬卷，南京库现货。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031205163), 9211, @now, 9211, @now, NULL),
  (9709, 'DEMO-P-9709', '无取向电工钢板卷 B50A470 0.50*1200*C', 'spot', '电工钢', '无取向电工钢板卷', '湛江钢铁',
   JSON_OBJECT('material','B50A470','spec','0.50*1200*C','thickness','0.50','width','1200','length','C','surface','A','edge','不切边','processing','电工钢卷'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-S6W0008797', 11.20, 11.20, 0.00, 4886.00, 1.00, 'CNY', 1, 13.00, 2, 9203, 'demo seed approved', @now,
   '/demo/certs/quality/S6W0008797.pdf', JSON_ARRAY('/demo/products/electrical-steel-coil.jpg'), '湛江无取向电工钢，适合电机与压缩机采购场景。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031204759), 9213, @now, 9213, @now, NULL),
  (9710, 'DEMO-P-9710', '热镀锌板卷 DC56D+Z 0.65*1630*C', 'spot', '热镀锌', '热镀锌板卷', '宝日汽车板',
   JSON_OBJECT('material','DC56D+Z','spec','0.65*1630*C','thickness','0.65','width','1630','length','C','surface','50/50 涂油 FD/M','edge','切边','processing','纯锌(Z)'),
   9502, 9302, JSON_OBJECT('warehouse_id',9502,'warehouse_service_id',9502,'warehouse_enterprise_id',9302,'warehouse_name','上海宝钢运输有限公司75号库(谭庄库)','warehouse_code','WS004693','address','上海市宝山区谭庄路75号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-M6W0065884', 1.32, 1.32, 0.00, 3856.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/M6W0065884.pdf', JSON_ARRAY('/demo/products/galvanized-coil.jpg'), '汽车板小吨位样品资源，适合试产。', 2, 48,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031355289), 9211, @now, 9211, @now, NULL),
  (9711, 'DEMO-P-9711', '热轧直发卷 SPHC-L 2.10*850*C', 'spot', '热轧', '热轧直发卷', '梅钢',
   JSON_OBJECT('material','SPHC-L','spec','2.10*850*C','thickness','2.10','width','850','length','C','surface','','edge','不切边','processing','热轧直发'),
   9503, 9303, JSON_OBJECT('warehouse_id',9503,'warehouse_service_id',9503,'warehouse_enterprise_id',9303,'warehouse_name','梅盛华康库','warehouse_code','WS000825','address','江苏省南京市浦口区梅盛华康库','province','江苏省','city','南京市','district','浦口区'),
   'OUYEEL-I6W0033322', 8.54, 8.54, 0.00, 3546.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/I6W0033322.pdf', JSON_ARRAY('/demo/products/hot-rolled-direct-coil.jpg'), '梅钢热轧直发卷，南京库。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031357655), 9211, @now, 9211, @now, NULL),
  (9712, 'DEMO-P-9712', '冷轧板卷 DC03 1.00*1320*C', 'spot', '冷轧', '冷轧板卷', '宝钢股份',
   JSON_OBJECT('material','DC03','spec','1.00*1320*C','thickness','1.00','width','1320','length','C','surface','FC/D','edge','切边','processing','简包装'),
   9501, 9301, JSON_OBJECT('warehouse_id',9501,'warehouse_service_id',9501,'warehouse_enterprise_id',9301,'warehouse_name','上海宝钢运输有限公司116号库(洪澳库)','warehouse_code','WS005593','address','上海市宝山区蕰川路116号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-L6W0032397', 7.41, 7.41, 0.00, 4226.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/L6W0032397.pdf', JSON_ARRAY('/demo/products/cold-rolled-coil.jpg'), '宝钢 DC03 冷轧现货。', 2, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031206154), 9211, @now, 9211, @now, NULL),
  (9713, 'DEMO-P-9713', '电镀锌板卷 DC06+ZE 0.70*1648*C', 'spot', '电镀锌', '电镀锌板卷', '宝钢股份',
   JSON_OBJECT('material','DC06+ZE','spec','0.70*1648*C','thickness','0.70','width','1648','length','C','surface','36/36 磷化+涂油','edge','切边','processing','纯锌(Z)'),
   9502, 9302, JSON_OBJECT('warehouse_id',9502,'warehouse_service_id',9502,'warehouse_enterprise_id',9302,'warehouse_name','上海宝钢运输有限公司75号库(谭庄库)','warehouse_code','WS004693','address','上海市宝山区谭庄路75号库','province','上海市','city','上海市','district','宝山区'),
   'OUYEEL-N6W0008924', 4.95, 4.95, 0.00, 4616.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/N6W0008924.pdf', JSON_ARRAY('/demo/products/electro-galvanized-coil.jpg'), '宝钢电镀锌尾板卷，EN 10152。', 2, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031207350), 9211, @now, 9211, @now, NULL),
  (9714, 'DEMO-P-9714', '热轧直发卷 Q345B 3.65*1160*C', 'spot', '热轧', '热轧直发卷', '梅钢',
   JSON_OBJECT('material','Q345B','spec','3.65*1160*C','thickness','3.65','width','1160','length','C','surface','','edge','不切边','processing','热轧直发'),
   9503, 9303, JSON_OBJECT('warehouse_id',9503,'warehouse_service_id',9503,'warehouse_enterprise_id',9303,'warehouse_name','梅盛华康库','warehouse_code','WS000825','address','江苏省南京市浦口区梅盛华康库','province','江苏省','city','南京市','district','浦口区'),
   'OUYEEL-I6W0029364', 9.41, 9.41, 0.00, 3446.00, 1.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/I6W0029364.pdf', JSON_ARRAY('/demo/products/hot-rolled-direct-coil.jpg'), '梅钢 Q345B 热轧直发卷。', 3, 72,
   JSON_OBJECT('source','ouyeel_sample','resource_id',2031398676), 9211, @now, 9211, @now, NULL),
  (9715, 'DEMO-P-9715', '热轧直发卷 SPHF-F 9.75*1440*C', 'spot', '热轧', '热轧直发卷', '梅钢',
   JSON_OBJECT('material','SPHF-F','spec','9.75*1440*C','thickness','9.75','width','1440','length','C','surface','','edge','不切边','processing','热轧直发'),
   9503, 9303, JSON_OBJECT('warehouse_id',9503,'warehouse_service_id',9503,'warehouse_enterprise_id',9303,'warehouse_name','梅盛华康库','warehouse_code','WS000825','address','江苏省南京市浦口区梅盛华康库','province','江苏省','city','南京市','district','浦口区'),
   'OUYEEL-TI15MG1027497', 97.38, 97.38, 0.00, 3356.00, 5.00, 'CNY', 1, 13.00, 2, 9201, 'demo seed approved', @now,
   '/demo/certs/quality/TI15MG1027497.pdf', JSON_ARRAY('/demo/products/hot-rolled-direct-coil.jpg'), '梅钢厚规格热轧直发卷，4件合计库存。', 3, 48,
   JSON_OBJECT('source','ouyeel_sample','sku_id',2568481685297308040,'total_count',4), 9211, @now, 9211, @now, NULL)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  category = VALUES(category),
  sub_category = VALUES(sub_category),
  steel_mill = VALUES(steel_mill),
  spec = VALUES(spec),
  warehouse_service_id = VALUES(warehouse_service_id),
  warehouse_enterprise_id = VALUES(warehouse_enterprise_id),
  warehouse_info = VALUES(warehouse_info),
  quantity = VALUES(quantity),
  available_qty = VALUES(available_qty),
  unit_price = VALUES(unit_price),
  status = VALUES(status),
  enterprise_id = VALUES(enterprise_id),
  approved_at = VALUES(approved_at),
  images = VALUES(images),
  description = VALUES(description),
  remark = VALUES(remark),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Warehouse operation data for warehouse dashboards.
INSERT INTO warehouse_inventory_movements (
  id, movement_no, warehouse_service_id, warehouse_enterprise_id, request_enterprise_id, product_id, product_name, bundle_no,
  order_id, order_no, movement_type, quantity, status, scheduled_at, completed_at, remark,
  created_by, updated_by, created_at, updated_at, deleted_at
) VALUES
  (9801, 'WM-DEMO-20260508-001', 9501, 9301, 9201, 9701, '冷轧板卷 HC340LA 2.00*1488*C', 'OUYEEL-L6W0029699', 0, '', 'inbound', 20.33, 'completed', '2026-05-08 09:30:00', @now, '欧冶样例资源入库验收完成', 9311, 9311, @now, @now, NULL),
  (9802, 'WM-DEMO-20260508-002', 9502, 9302, 9201, 9704, '热镀锌板卷 DC51D+Z 0.65*1600*C', 'OUYEEL-TL21BG1001576', 0, '', 'inbound', 22.98, 'completed', '2026-05-08 10:00:00', @now, '平台好货入库', 9312, 9312, @now, @now, NULL),
  (9803, 'WM-DEMO-20260508-003', 9503, 9303, 9201, 9715, '热轧直发卷 SPHF-F 9.75*1440*C', 'OUYEEL-TI15MG1027497', 0, '', 'inbound', 97.38, 'completed', '2026-05-08 11:00:00', @now, '梅盛华康库大货入库', 9313, 9313, @now, @now, NULL),
  (9804, 'WM-DEMO-20260508-004', 9501, 9301, 9101, 9706, '冷轧板卷 BLD 1.00*1075*C', 'OUYEEL-L6W0032376', 0, 'ORD-DEMO-20260508-001', 'outbound', 3.00, 'processing', '2026-05-09 09:30:00', NULL, '买方预约出库，待装车', 9311, 9311, @now, @now, NULL)
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  scheduled_at = VALUES(scheduled_at),
  completed_at = VALUES(completed_at),
  remark = VALUES(remark),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

-- Logistics operation data for logistics dashboards.
INSERT INTO logistics_orders (
  id, logistics_no, order_id, order_no, buyer_enterprise_id, logistics_service_id, carrier_enterprise_id,
  warehouse_service_id, origin, destination, cargo_name, weight, vehicle_type, estimated_price, final_price,
  status, driver_name, driver_phone, plate_number, guarantee_type, accepted_by, accepted_at,
  created_by, updated_by, created_at, updated_at, deleted_at
) VALUES
  (9901, 'LG-DEMO-20260508-001', 0, 'ORD-DEMO-20260508-001', 9101, 9601, 9401, 9501,
   '上海宝钢运输有限公司116号库(洪澳库)', '上海市嘉定区安亭镇墨玉南路888号', '冷轧板卷 BLD', 3.00, '13米半挂/卷钢鞍座车', 1250.00, 1250.00,
   'accepted', '李成', '13611112222', '沪A8D521', '平台承运险', 9411, @now, 9111, 9411, @now, @now, NULL),
  (9902, 'LG-DEMO-20260508-002', 0, 'ORD-DEMO-20260508-002', 9102, 9602, 9402, 9503,
   '梅盛华康库', '湖北省武汉市青山区和平大道999号', '热轧直发卷 Q345B', 9.41, '17.5米平板', 3200.00, 3180.00,
   'in_transit', '周勇', '13633334444', '苏A6K902', '平台承运险', 9412, @now, 9112, 9412, @now, @now, NULL),
  (9903, 'LG-DEMO-20260508-003', 0, 'ORD-DEMO-20260508-003', 9101, 9603, 9403, 9501,
   '湛江宝钢高强钢科技有限公司', '广东省佛山市顺德区北滘镇工业大道', '热镀锌板卷 DC54D+ZF', 4.12, '13米半挂/卷钢专用车', 910.00, 910.00,
   'delivered', '何军', '13655556666', '粤G7P318', '货损险', 9413, @now, 9111, 9413, @now, @now, NULL)
ON DUPLICATE KEY UPDATE
  status = VALUES(status),
  driver_name = VALUES(driver_name),
  driver_phone = VALUES(driver_phone),
  plate_number = VALUES(plate_number),
  final_price = VALUES(final_price),
  updated_at = VALUES(updated_at),
  deleted_at = NULL;

INSERT INTO logistics_tracking_events (
  id, logistics_order_id, event_type, location, description, occurred_at, created_by, created_at
) VALUES
  (9911, 9901, 'accepted', '上海市宝山区', '承运商已接单，车辆已分配。', @now, 9411, @now),
  (9912, 9902, 'picked_up', '南京市浦口区', '车辆已在梅盛华康库提货。', DATE_SUB(@now, INTERVAL 3 HOUR), 9412, @now),
  (9913, 9902, 'in_transit', '沪蓉高速南京段', '车辆在途，预计明日到达武汉。', @now, 9412, @now),
  (9914, 9903, 'delivered', '广东省佛山市顺德区', '货物已送达买方指定工厂。', @now, 9413, @now)
ON DUPLICATE KEY UPDATE
  logistics_order_id = VALUES(logistics_order_id),
  event_type = VALUES(event_type),
  location = VALUES(location),
  description = VALUES(description),
  occurred_at = VALUES(occurred_at),
  created_by = VALUES(created_by);

SET FOREIGN_KEY_CHECKS = 1;
