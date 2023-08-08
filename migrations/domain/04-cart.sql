CREATE TABLE `cart` (
  `id` VARCHAR(55) PRIMARY KEY NOT NULL,
  `user_id` VARCHAR(55) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
  FOREIGN KEY (`user_id`) REFERENCES `User`(`user_id`)
);

CREATE TABLE `cart_item` (
  `cart_id` VARCHAR(55) NOT NULL,
  `product_id` VARCHAR(55) NOT NULL,
  `quantity` INT NOT NULL,
  `cost` DECIMAL(10,2) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
  FOREIGN KEY (`cart_id`) REFERENCES `Cart`(`id`),
  FOREIGN KEY (`product_id`) REFERENCES `Product`(`id`)
);

CREATE TABLE `order` (
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
  FOREIGN KEY (`user_id`) REFERENCES `User`(`user_id`)
);

CREATE TABLE `order_item` (
  `order_id` VARCHAR(55) NOT NULL,
  `product_id` VARCHAR(55) NOT NULL,
  `quantity` INT NOT NULL,
  `cost` DECIMAL(10,2),
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
  FOREIGN KEY (`order_id`) REFERENCES `Order`(`id`),
  FOREIGN KEY (`product_id`) REFERENCES `Product`(`id`)
);

INSERT INTO `cart` (`id`, `user_id`, `created_by`)
VALUES 
('990e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'john');

INSERT INTO `cart_item` (`cart_id`, `product_id`, `quantity`, `cost`, `created_by`)
VALUES 
('990e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 1, 799.00, 'john'),
('990e8400-e29b-41d4-a716-446655440000', '770e8400-e29b-41d4-a716-446655440000', 2, 1398.00, 'john');

INSERT INTO `order` (`id`, `user_id`, `total_cost`, `status`, `created_by`)
VALUES 
('cc0e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 2197.00, 'processing', 'john');

INSERT INTO `order_item` (`order_id`, `product_id`, `quantity`, `cost`, `created_by`)
VALUES 
('cc0e8400-e29b-41d4-a716-446655440000', '660e8400-e29b-41d4-a716-446655440000', 1, 799.00, 'john');