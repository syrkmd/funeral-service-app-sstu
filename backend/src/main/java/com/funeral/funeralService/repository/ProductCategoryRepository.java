package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.ProductCategory;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface ProductCategoryRepository extends JpaRepository<ProductCategory, Long> {

    List<ProductCategory> findByActiveTrueOrderBySortOrderAsc();
}
