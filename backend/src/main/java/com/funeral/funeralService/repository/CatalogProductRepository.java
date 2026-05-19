package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.CatalogProduct;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface CatalogProductRepository extends JpaRepository<CatalogProduct, Long> {

    List<CatalogProduct> findByActiveTrueOrderBySortOrderAsc();

    List<CatalogProduct> findByCategoryIdAndActiveTrueOrderBySortOrderAsc(Long categoryId);
}
