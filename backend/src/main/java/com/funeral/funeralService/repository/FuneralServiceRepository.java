package com.funeral.funeralService.repository;

import com.funeral.funeralService.entity.FuneralService;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface FuneralServiceRepository extends JpaRepository<FuneralService, Long> {
}
