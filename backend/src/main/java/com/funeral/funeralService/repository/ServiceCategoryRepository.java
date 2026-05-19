package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.FuneralService;
import com.funeral.funeralService.entity.ServiceCategory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ServiceCategoryRepository extends JpaRepository<ServiceCategory, Long> {

    List<ServiceCategory> findByActiveTrueOrderBySortOrderAsc();
}
