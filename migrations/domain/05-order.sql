CREATE TABLE `orders` (
  `id` VARCHAR(55) PRIMARY KEY,
  `user_id` VARCHAR(55),
  `total_cost` DECIMAL(10,2),
  `status` ENUM('pending', 'processing', 'shipped', 'delivered', 'canceled'),
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
);

CREATE TABLE `order_item` (
  `order_id` VARCHAR(55) NOT NULL,
  `product_id` VARCHAR(55) NOT NULL,
  `unit_price` DECIMAL(10,2) NOT NULL,
  `quantity` INT NOT NULL,
  `cost` DECIMAL(10,2),
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL,
  CONSTRAINT `fk_order_item_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`),
  CONSTRAINT `fk_order_item_product` FOREIGN KEY (`product_id`) REFERENCES `product`(`id`)
);