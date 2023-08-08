DROP TABLE IF EXISTS `product`;

CREATE TABLE `product` (
  `id` VARCHAR(55) PRIMARY KEY NOT NULL,
  `user_id` VARCHAR(55) NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `brand` VARCHAR(255) NOT NULL,
  `category` VARCHAR(255) NOT NULL,
  `stock` INT NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
);

INSERT INTO `Product` (`id`, `name`, `price`, `brand`, `category`, `stock`, `created_by`)
VALUES 
('660e8400-e29b-41d4-a716-446655440000', 'iPhone 13', 799.00, 'Apple', 'Electronics', 500, 'admin'),
('770e8400-e29b-41d4-a716-446655440000', 'Galaxy S21', 699.00, 'Samsung', 'Electronics', 400, 'admin'),
('880e8400-e29b-41d4-a716-446655440000', 'Macbook Pro', 1299.00, 'Apple', 'Computers', 300, 'admin');