package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.CatalogProduct;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface CatalogProductRepository extends JpaRepository<CatalogProduct, Long> {

    List<CatalogProduct> findByActiveTrueOrderBySortOrderAsc();

    List<CatalogProduct> findAllByOrderBySortOrderAsc();

    List<CatalogProduct> findByCategoryIdAndActiveTrueOrderBySortOrderAsc(Long categoryId);

    List<CatalogProduct> findByCategoryIdOrderBySortOrderAsc(Long categoryId);

    Optional<CatalogProduct> findFirstByTitleAndActiveTrue(String title);
}
