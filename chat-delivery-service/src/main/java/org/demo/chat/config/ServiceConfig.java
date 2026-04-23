package org.demo.chat.config;

import io.smallrye.config.ConfigMapping;

@ConfigMapping(prefix = "service")
public interface ServiceConfig {
	
	String messageServiceUrl();
	
	String jwtSecret();
	String jwtIssuer();

}
