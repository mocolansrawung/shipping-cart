CREATE TABLE `cart` (
  `id` VARCHAR(55) PRIMARY KEY NOT NULL,
  `user_id` VARCHAR(55) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL
);

CREATE TABLE `cart_item` (
  `cart_id` VARCHAR(55) NOT NULL,
  `product_id` VARCHAR(55) NOT NULL,
  `unit_price` DECIMAL(10,2) NOT NULL,
  `quantity` INT NOT NULL,
  `cost` DECIMAL(10,2) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_by` VARCHAR(55) NOT NULL,
  `updated_at` TIMESTAMP NULL DEFAULT NULL,
  `updated_by` VARCHAR(55) NULL DEFAULT NULL,
  `deleted_at` TIMESTAMP NULL DEFAULT NULL,
  `deleted_by` VARCHAR(55) NULL DEFAULT NULL,
  CONSTRAINT `fk_cart_item_cart` FOREIGN KEY (`cart_id`) REFERENCES `cart`(`id`),
  CONSTRAINT `fk_cart_item_product` FOREIGN KEY (`product_id`) REFERENCES `product`(`id`)
);

-- Insert a cart for the user
INSERT INTO `cart` (`id`, `user_id`, `created_by`)
VALUES ('c1234567-89ab-cdef-0123-456789abcdef', '2407ae50-3a74-49f9-876c-ecef1087229f', 'syst2407ae50-3a74-49f9-876c-ecef1087229fem');

-- Insert multiple items into the cart
-- Assuming the product IDs: 'p1', 'p2', 'p3' exist in the `product` table
INSERT INTO `cart_item` (`cart_id`, `product_id`, `price`, `quantity`, `cost`, `created_by`)
VALUES 
('c1234567-89ab-cdef-0123-456789abcdef', '1c0a7155-f770-4c64-ad9e-6ca5d12b2636', 99.99, 1, 99.99, '2407ae50-3a74-49f9-876c-ecef1087229f'),
('c1234567-89ab-cdef-0123-456789abcdef', 'e11f94c4-60b1-4da6-be31-379b1ff2a117', 99.99, 1, 99.99, '2407ae50-3a74-49f9-876c-ecef1087229f');

DELIMITER //

CREATE TRIGGER update_cart_item_price
AFTER UPDATE ON product
FOR EACH ROW 
BEGIN
    IF OLD.price != NEW.price THEN
        UPDATE cart_item
        SET unit_price = NEW.price
        WHERE product_id = NEW.id;
    END IF;
END;

//

DELIMITER ;

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